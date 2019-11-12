# Go RemDVC

 [![GoDoc][doc-img]][doc]

*Remote Embedded Distributed Version Control...* say what?

Save internal configuration for your application in git using go. Allowing your clients to save their configuration in a local git version control library. We have libaries for easy interagtion to HTTP and SSH servers.

## What does this library do

It provides built-in HTTP and SSH handlers for running an embeded git server that provides pre-receive and post-receive hooks within your application. One reason you may want to do this is to provide a way for users to use git to push config changes to your application, which is exactly what we do for XA4B (http://xa4b.com). We use this library and the post-receive hook to publish the configuration to other connected remote instances. Using Git as the method in which to manage confiuration throughout a distrubted system without a single source database.

## Installation

Go 1.13 or later is required

`go get -u gopkg.xa4b.com/git/cfghttp`

or

`go get -u gopkg.xa4b.com/git/cfgssh`

## Quick Start

Using the `cfghttp` library to quickly add a git configuration.

```go
// ...
import "gopkg.xa4b.com/git/cfg"
import "gopkg.xa4b.com/git/cfghttp"
// ...

type configuration struct {
    ExampleField string 
}

func main() {
    repos := make(map[string]*gogit.Repository)
    repos["/config"] = git.Init(memory.Store, nil) // in-memory store

    configReload := make(chan struct{}, 1)

    go func() {
        conf := &configuration{}
        loadConfiguration(conf) // load from a file or defaults

        // run a http server to access your git file with the following: 
        //
        //    git clone localhost:8333/internal/config
        //
        http.ListenAndServe(":8333", cfghttp.NewServer(
            cfghttp.LoadGoGit(repos, "internal"),

            // once you push a new config with: git push
            // the update and reload
            cfghttp.WithPostReceiveHook(func(w io.Writer, data *cfg.PostReceivePackHookData){
                for _, d := range data.Refs {
                    // uh, check for errors below
                    obj, _ := repos[data.RepoName].CommitObject(plumbing.NewHash(v.NewHash))
			        f, _ := obj.File("configuration.toml")
                    r, _ := f.Blob.Reader()
                    if isValidCommit(r) {
                        configReload <- struct{}{} // kick off reload
                         fmt.Fprint(w, "the config file was reloaded, congrats.")
                        return nil
                    }
                    fmt.Fprint(w, "the config file was\n\nnot\n\nreloaded.")
                    return
                }
                fmt.Fprintln(w, "go new config, and applying")
            }),
        ))
    }()

    for {
        // ... load app and run
       do := <- captureSignalsAndReloadConfiguration(configReload) // block until config reload or a ^C signal
       if do == "exit" {
           os.Exit(0)
       }
    }
}
```

## Development Status: Alpha

There are no plans to drasticly change the API, but we are leaving the libray in *Alpha* status until there has been more usage.

## Contributing

We support an active community of contributors &mdash; and yeah that includes you! 
Feel free to fork and submit pull requests for any updates or impovements. We are
gathering information for Community and Contribution guidelines. As of now please
mantain the golden rule while interating with this project; which is most
familarly: *“Do unto others as you would have them do unto you.”

<hr>

Released under the [MIT License](LICENSE.txt).

[doc-img]: https://godoc.org/gopkg.xa4b.com/git?status.svg
[doc]: https://godoc.org/gopkg.xa4b.com/git
