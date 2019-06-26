# Schwer

_adjective_ | /ʃveːɐ/ | _meaning_: heavy

A single binary application with built-in web front-end that can produce cpu/mem load.


## Installation

1. Clone this repo
2. Change working directory to your local cloned repo
3. `$ go generate . && go build .`


## Usage

Once you built the application, just run it. Schwer, by default, binds to port `9999`, but you
can specify a different port (between 1024-65535) if you like by using the `-port` flag:

`$ ./schwer -port 19999`


### Web

While Schwer is running, you can open the web front-end in your browser by visiting `localhost:<port>`.

You should see something like this:

![Schwer index page](img/schwer_index.png)

The amount of CPU load percentage Schwer can produce can be specified by typing a number between
0 - 100 into the little input field and hitting enter or clicking the `[Update]` button.


### API

You can also use Schwer via its HTTP API.

| Endpoint | Method | Params | Response code |  Description |
| -------- | ------ | ------ | ------------- |  ----------- |
| `/cpu`   | `GET`  | `-`    | 200 OK        | Returns an array of CPU utilisation levels per core (e.g. `[49, 34, 50, 32]` in case of a machine with 4 cores). |
| `/cpu`   | `POST` | `pct` - load level % (0-100) | 202 Accepted<br>400 Bad Request | Sets the load level for Schwer to produce. |


## TODO

- [ ] Flesh out CPU load
- [x] CPU load monitor
- [ ] Memory load
- [ ] Memory load monitor
- [ ] Make it work on Windows
