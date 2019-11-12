package cfghttp

// This file wraps around the go-git library to create a GitServer interface

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	logg "log"
	"net/http"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/capability"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/server"
	"gopkg.xa4b.com/git/cfg"
	"gopkg.xa4b.com/git/pktline"
)

// LoadGoGit loads a mapping of git repositories (go-git) to a repository endpoint
func LoadGoGit(m map[string]*git.Repository, endpoint string) *GoGitServer {
	ml := make(server.MapLoader)
	for k, v := range m {
		ml[endpoint+k] = v.Storer
	}

	caps := []capability.Capability{
		capability.Sideband,
		capability.Sideband64k,
		capability.PushOptions,
	}

	return &GoGitServer{repos: m, transport: server.NewServer(ml), endpoint: endpoint, capabilities: caps, log: log{}}
}

// GoGitServer wraps concepts for go-git into a GitServer HTTP interface
type GoGitServer struct {
	repos     map[string]*git.Repository
	transport transport.Transport
	endpoint  string

	capabilities []capability.Capability
	log          log

	preReceiveHookFn  cfg.PreReceivePackHookFunc
	postReceiveHookfn cfg.PostReceivePackHookFunc
}

// WithLogger takes in logger/s to display debug and info logs for the GoGitServer object
func (s *GoGitServer) WithLogger(logger ...interface{}) {
	for _, l := range logger {
		switch v := l.(type) {
		case DebugLogger:
			s.log.debug = v
		case InfoLogger:
			s.log.info = v
		case *logg.Logger:
			s.log.debug, s.log.info = v, v
		default:
			logg.Printf("invalid logger %T passed", v)
		}
	}
}

// WithPreReceiveHook sets the pre-receive hook for receive-pack requests
func (s *GoGitServer) WithPreReceiveHook(fn cfg.PreReceivePackHookFunc) {
	s.preReceiveHookFn = fn
}

// WithPostReceiveHook sets the post-receive hook for receive-pack requests
func (s *GoGitServer) WithPostReceiveHook(fn cfg.PostReceivePackHookFunc) {
	s.postReceiveHookfn = fn
}

// InfoRefs holds all of the data needed to handle the git interface for
// info-ref requests
type InfoRefs struct {
	*GoGitServer
	repoName string

	service string
	refs    *packp.AdvRefs

	logPrefix string
	err       error
}

// NewInfoRefs returns a new InfoRefs object to handle InfoRefs requests
func (s *GoGitServer) NewInfoRefs(repoName string) InfoRefser {
	s.log.Debug("fn: NewInfoRefs...")
	return &InfoRefs{GoGitServer: s, repoName: repoName, logPrefix: "info-refs:"}
}

// DoHTTP takes in a HTTP request and decodes what is supposed to happen
// if there are any errors then this method is skipped, and errors can be checked with
// the Err() method
func (ir *InfoRefs) DoHTTP(w http.ResponseWriter, r *http.Request) InfoRefser {
	ir.log.Debug(ir.logPrefix, "fn: DecodeHTTP...")

	if ir.err != nil {
		ir.log.Debug(ir.logPrefix, "skip: on error")
		return ir
	}

	var q = r.URL.Query()
	if serv, ok := q["service"]; ok {
		if len(serv) > 1 {
			ir.err = ErrNoServiceFound
		}
		ir.service = serv[0]
	}

	endpoint, err := transport.NewEndpoint(ir.endpoint + ir.repoName)
	if err != nil {
		return ir.withErr(ErrTransportEndpoint.F(ir.repoName, err))
	}

	switch ir.service {
	case "git-receive-pack":
		if rps, err := ir.transport.NewReceivePackSession(endpoint, nil); err != nil {
			return ir.withErr(ErrSession.F(ir.service, err))
		} else if ir.refs, err = rps.AdvertisedReferences(); err != nil {
			return ir.withErr(ErrSessionAdvRefs.F(ir.service, err))
		}
	case "git-upload-pack":
		if ups, err := ir.transport.NewUploadPackSession(endpoint, nil); err != nil {
			return ir.withErr(ErrSession.F(ir.service, err))
		} else if ir.refs, err = ups.AdvertisedReferences(); err != nil {
			return ir.withErr(ErrSessionAdvRefs.F(ir.service, err))
		}
	}

	for _, cap := range ir.capabilities {
		ir.refs.Capabilities.Set(cap)
	}

	ir.log.Info(ir.logPrefix, "setting the proper headers")
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-%s-advertisement", ir.service))
	w.Header().Set("Cache-Control", "no-cache")

	enc := pktline.NewEncoder(w)
	enc.EncodeString(fmt.Sprintf("# service=%s\n", ir.service))
	enc.Flush()

	ir.log.Info(ir.logPrefix, "sending back the proper references...")
	ir.refs.Encode(w)

	return ir
}

// Err return any errors
func (ir *InfoRefs) Err() error { return ir.err }

// withErr sets the object error and returns the object
// so that method chaining will work as expected.
func (ir *InfoRefs) withErr(err error) InfoRefser {
	ir.log.Info(ir.logPrefix, "ERR:", err)
	ir.err = err
	return ir
}

// ReceivePack holds all of the data needed to handle the git interface for
// receive-pack requests through HTTP
type ReceivePack struct {
	*GoGitServer

	repoName string
	encOpts  []pktline.EncoderOption
	cleanup  []func()

	sess  transport.ReceivePackSession
	refs  *packp.AdvRefs
	rReq  *packp.ReferenceUpdateRequest
	rStat *packp.ReportStatus

	logPrefix string
	err       error
}

// NewReceivePack returns a new ReceivePack object
func (s *GoGitServer) NewReceivePack(repoName string) ReceivePacker {
	return &ReceivePack{
		GoGitServer: s,
		repoName:    repoName,
		logPrefix:   "receive-pack [HTTP]:",
	}
}

// DoHTTP takes in a HTTP request and decodes what is supposed to happen
// if there are any errors then this method is skipped, and errors can be checked with
// the Err() method
func (rp *ReceivePack) DoHTTP(w http.ResponseWriter, r *http.Request) ReceivePacker {
	rp.log.Debug(rp.logPrefix, "fn: DoHTTP...")

	if rp.err != nil {
		rp.log.Debug(rp.logPrefix, "skip: on error")
		return rp
	}

	endpoint, err := transport.NewEndpoint(rp.endpoint + rp.repoName)
	if err != nil {
		return rp.withErr(ErrTransportEndpoint.F(rp.repoName, err))
	}

	if rp.sess, err = rp.transport.NewReceivePackSession(endpoint, nil); err != nil {
		return rp.withErr(ErrSession.F("receive-pack", err))
	}

	if rp.refs, err = rp.sess.AdvertisedReferences(); err != nil {
		return rp.withErr(ErrSessionAdvRefs.F("receive-pack", err))
	}

	for _, cap := range rp.capabilities {
		rp.refs.Capabilities.Set(cap)
	}

	rp.rReq = packp.NewReferenceUpdateRequest()

	buf := new(bytes.Buffer)
	if err = rp.rReq.Decode(io.TeeReader(r.Body, buf)); err != nil {
		return rp.withErr(ErrPackDecode.F(err))
	}
	rp.addCleanup(func() { r.Body.Close() }) // always close the body

	var hookData *cfg.ReceivePackHookData
	// TODO(njones): add push-option support
	rp.encOpts, hookData, err = encOpts(strings.NewReader(buf.String()))
	if err != nil {
		return rp.withErr(err) // is a pre-wrapped error
	}
	hookData.RepoName = rp.repoName

	rp.log.Info("adding the pre-receive hook...")
	// the git pre-receive-hook function
	if rp.preReceiveHookFn != nil {
		rp.log.Debug(rp.logPrefix, "fn: (preHookFn)...")
		buf := new(bytes.Buffer)
		if hookData == nil {
			return rp.withErr(ErrEmptyHookData.F(err))
		}
		refBranch, err := rp.preReceiveHookFn(buf, &cfg.PreReceivePackHookData{*hookData})
		if err != nil {
			enc := pktline.NewEncoder(w, rp.encOpts...).WithSidebandCapability(pktline.Sideband64k)
			scn := bufio.NewScanner(buf)
			for scn.Scan() {
				enc.Sideband.EncodeProgress(scn.Text() + "\n")
			}
			enc.EncodeString("unpack ok\n")
			enc.EncodeString(fmt.Sprintf("ng %s %s\n", refBranch, err))
			enc.Sideband.Flush()
			enc.Flush()

			return rp // done
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	rp.addCleanup(func() { cancel() })

	if rp.rStat, err = rp.sess.ReceivePack(ctx, rp.rReq); err != nil {
		return rp.withErr(ErrReceivePack.F(err))
	}

	if rp.postReceiveHookfn != nil {
		if hookData == nil {
			return rp.withErr(ErrEmptyHookData.F(err))
		}
		pr, pw := io.Pipe()
		go func() {
			rp.postReceiveHookfn(pw, &cfg.PostReceivePackHookData{*hookData, make(map[string]string)})
			pw.Close()
		}()
		enc := pktline.NewEncoder(w, rp.encOpts...).WithSidebandCapability(pktline.Sideband64k)
		scn := bufio.NewScanner(pr)
		for scn.Scan() {
			enc.Sideband.EncodeProgress(scn.Text() + "\n")
		}
	}

	enc := pktline.NewEncoder(w, rp.encOpts...).WithSidebandCapability(pktline.Sideband64k)
	pktline.ReportStatus(enc, rp.rStat)
	enc.Flush()

	return rp
}

// Cleanup takes any functions that were collected and runs them. This is for
// deferred processes
func (rp *ReceivePack) Cleanup() {
	rp.log.Debug(rp.logPrefix, "fn: Cleanup...")

	for _, fn := range rp.cleanup {
		rp.log.Info(rp.logPrefix, "cleaning up all of the processes")
		fn()
	}
}

// Err returns the error that was collected during the receive-pack processing
func (rp *ReceivePack) Err() error {
	rp.log.Debug(rp.logPrefix, "fn: Err...")

	return rp.err
}

// withErr sets the object err field and returns the object so
// that object chaining will work as expected.
func (rp *ReceivePack) withErr(err error) ReceivePacker {
	rp.log.Debug(rp.logPrefix, "fn: withErr...")
	rp.log.Info(rp.logPrefix, "ERR:", err)
	rp.err = err
	return rp
}

// addCleanup simply appends a function to an arry so that cleanup can
// happen after all of the processing has been done.
func (rp *ReceivePack) addCleanup(fn func()) {
	rp.log.Debug(rp.logPrefix, "fn: addCleanup...")
	rp.cleanup = append(rp.cleanup, fn)
}

// UploadPack holds all of the data needed to handle the git interface for
// upload-pack requests for HTTP
type UploadPack struct {
	*GoGitServer
	repoName string

	sess  transport.UploadPackSession
	refs  *packp.AdvRefs
	uReq  *packp.UploadPackRequest
	uResp *packp.UploadPackResponse

	cleanup []func()

	logPrefix string
	err       error
}

// NewUploadPack returns a new UploadPack object
func (s *GoGitServer) NewUploadPack(repoName string) UploadPacker {
	return &UploadPack{GoGitServer: s, repoName: repoName, logPrefix: "upload-pack [HTTP]:"}
}

// DoHTTP takes in a HTTP request and decodes what is supposed to happen
// if there are any errors then this method is skipped, and errors can be checked with
// the Err() method
func (up *UploadPack) DoHTTP(w http.ResponseWriter, r *http.Request) UploadPacker {
	up.log.Debug(up.logPrefix, "DoHTTP...")

	if up.err != nil {
		up.log.Debug(up.logPrefix, "skip: on error")
		return up
	}

	endpoint, err := transport.NewEndpoint(up.endpoint + up.repoName)
	if err != nil {
		return up.withErr(ErrTransportEndpoint.F(up.repoName, err))
	}

	if up.sess, err = up.transport.NewUploadPackSession(endpoint, nil); err != nil {
		return up.withErr(ErrSession.F("upload-pack", err))
	}

	up.uReq = packp.NewUploadPackRequest()
	if err = up.uReq.Decode(r.Body); err != nil {
		return up.withErr(ErrUploadPackRequest.F(err))
	}
	defer r.Body.Close() // close when we're done

	// the capablities need to be added before the UploadPack call
	if up.refs, err = up.sess.AdvertisedReferences(); err != nil {
		return up.withErr(ErrSessionAdvRefs.F("upload-pack", err))
	}

	for _, cap := range up.capabilities {
		up.refs.Capabilities.Set(cap)
	}

	ctx, cancel := context.WithCancel(context.Background())
	up.addCleanup(func() { cancel() }) // cancel when we are finished consuming integers

	up.uResp, err = up.sess.UploadPack(ctx, up.uReq)
	if err != nil {
		return up.withErr(ErrUploadPack.F(err))
	}

	// buffer the upload pack response
	buf := new(bytes.Buffer)
	up.uResp.Encode(buf) // this doesn't work with the go-git sideband muxer

	// re-write out the pack data with side-band awareness
	enc := pktline.NewEncoder(w, pktline.WithSideband64kMuxer).WithSidebandCapability(pktline.Sideband64k)
	enc.Write(buf.Bytes()[:8]) // already encoded...
	enc.WriteString(fmt.Sprintf("%04x\x01", 1+4+len(buf.Bytes()[8:])))
	enc.Write(buf.Bytes()[8:])
	enc.Flush()

	return up
}

// Cleanup takes any functions that were collected and runs them. This is for
// deferred processes
func (up *UploadPack) Cleanup() {
	up.log.Debug(up.logPrefix, "fn: Cleanup...")
	for _, fn := range up.cleanup {
		fn()
	}
}

// Err returns the error that was collected during the upload-pack processing
func (up *UploadPack) Err() error {
	up.log.Debug(up.logPrefix, "fn: Err...")
	return up.err
}

// withErr sets the object err field and returns the object so
// that object chaining will work as expected.
func (up *UploadPack) withErr(err error) UploadPacker {
	up.log.Debug(up.logPrefix, "fn: withErr...")
	up.log.Info(up.logPrefix, "ERR:", err)
	up.err = err
	return up
}

// addCleanup simply appends a function to an arry so that cleanup can
// happen after all of the processing has been done.
func (up *UploadPack) addCleanup(fn func()) {
	up.log.Debug(up.logPrefix, "fn: addCleanup...")
	up.cleanup = append(up.cleanup, fn)
}

// encOpts returns the encoding options, the hookdata and any errors
// from the receive-pack upload and will be used for the (pre|post)-receive pack hook
func encOpts(r io.Reader) ([]pktline.EncoderOption, *cfg.ReceivePackHookData, error) {
	var oldHash, newHash, refName string
	var encOpts = make([]pktline.EncoderOption, 0, 1)
	var hookData = &cfg.ReceivePackHookData{Refs: make([]cfg.ReceivePackData, 0)}
	var skip = false

	scn := pktline.NewScanner(r)
	for scn.Scan() {
		// looking for something like: a6a63a... 6684c0... refs/heads/master\0 report-status side-band-64k push-options agent=git/2.20.1
		refCapSplit := strings.Split(scn.Text(), "\x00")
		if len(refCapSplit) == 2 {
			_, err := fmt.Sscan(refCapSplit[0], &oldHash, &newHash, &refName)
			if err != nil {
				return nil, nil, ErrPackScanAdvRefs.F(err)
			}
			hookData.Refs = append(hookData.Refs, cfg.ReceivePackData{OldHash: oldHash, NewHash: newHash, RefName: refName})
		}
		if !skip {
			if len(refCapSplit) > 1 {
				idx := strings.Index(refCapSplit[1], "side-band")
				if idx > -1 {
					if refCapSplit[1][idx+9:][:4] == "-64k" {
						encOpts = append(encOpts, pktline.WithSideband64kMuxer)
					} else {
						encOpts = append(encOpts, pktline.WithSidebandMuxer)
					}
				}
			}
			skip = true
		}
	}

	return encOpts, hookData, nil
}
