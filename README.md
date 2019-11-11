# The GitConfater

 [![GoDoc][doc-img]][doc]

Save internal configuration for your application in git using go. Allowing your clients to save their configuration in a local git version control library. We have libraries for easy integration to HTTP and SSH servers.

## Installation

`go get -u gopkg.xa4b.com/git`

Go 1.13 or later is required

## Quick Start

Using the `githttp` library to quickly add a git configuration.

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

There are no plans to drastically change the API, but we are leaving the library in *Alpha* status until there has been more usage.

## Contributing

We support an active community of contributors &mdash; and yeah that includes you! 
Feel free to fork and submit pull requests for any updates or improvements. We are
gathering information for Community and Contribution guidelines. As of now please
maintain the golden rule while participating in this project; which is most
familiarly: *“Do unto others as you would have them do unto you.”

<hr>

Released under the [MIT License](LICENSE.txt).

[doc-img]: https://godoc.org/gopkg.xa4b.com/git?status.svg
[doc]: https://godoc.org/gopkg.xa4b.com/git
