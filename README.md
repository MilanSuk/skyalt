<p align="center">
<img src="https://raw.githubusercontent.com/MilanSuk/skyalt/main/screenshots/logo.png" />
</p>

<br/>
<p align="center">
<img src="https://raw.githubusercontent.com/MilanSuk/skyalt/main/screenshots/main.gif" style="border:1px solid LightGrey" />
</p>

[More screenshots here](https://github.com/MilanSuk/skyalt/tree/main/screenshots)


# SkyAlt
Build local-first apps on top of SQLite files.

Web: https://www.skyalt.com/

Twitter: https://twitter.com/skyalt



## From SaaS to Local-first software
Most of today's apps run in a browser as Software as a Service. Here's the list of problems you may experience:
- delay between client and server
- none or simple export
- hard to migrate between clouds
- data disappear(music playlist, etc.)
- data was tampered
- new SaaS version was released and you wanna keep using the older one
- no offline mode
- SaaS shut down
- price goes up
- 3rd party can access your data
<br/><br/>

SkyAlt solves them with [Local-first software](https://www.inkandswitch.com/local-first/)([video](https://www.youtube.com/watch?v=KrPsyr8Ig6M)). The biggest advantages can be summarized as:
- quick responses
- works offline
- ownership
- privacy(E2EE everywhere)
- works 'forever' + run any version



## From Webkit to WASM
There are few implementations of local-first platforms and most of them use Webkit. Webkit is huge and browsers are most complex things humans build and maintain. SkyAlt is heading in oposite direction - build large and complex worlds with few simple tools.

**Front-end**: Instead of writing app in HTML/CSS/JS, you pick up from many languages which compile to WASM and you use SkyAlt's apis() to draw on screen through [Immediate mode GUI](https://en.wikipedia.org/wiki/Immediate_mode_GUI) model.

**Back-end**: There is no back-end. Front-end uses SQL to read/write data from local SQLite files.

**Debugging**: The best tools to write and debug code are the ones developers already use. Every SkyAlt app can be compile into WASM *or* can be run as binary in separate process, which connects to SkyAlt and communicate over TCP socket. That means that developer can use any IDE and debugger, iterate quickly and compile app into wasm for final shipping.

**Formats**: For endurance, SkyAlt uses only well-known and open formats:
- WASM for binaries
- SQLite for storages
- Json for settings



## App examples
- [7Gui](https://github.com/milansuk/skyalt/blob/main/apps/7gui/main.go)
- [Calendar](https://github.com/milansuk/skyalt/blob/main/apps/calendar/main.go)
- [Map](https://github.com/milansuk/skyalt/blob/main/apps/map/main.go)
- [Database](https://github.com/milansuk/skyalt/blob/main/apps/db/main.go)



## Current state
- **Experimental**
- go-lang SDK only
- Linux / Windows / Mac(untested)
- ~16K LOC

*SkyAlt is ~3 months old. Right now, highest priority is providing best developer experience through high range of use-cases so we iterate and change apis() a lot => apps need to be edited and recompiled to wasm!*



## Versions
- v0.0(Aug 5, 2023): 1st line written.
- v0.1(Sep 11, 2023): Basic demo. Calendar, Map, 7Gui apps.
- v0.2(Sep 25, 2023): Styles(for buttons, texts, etc.). Better debugging. Performance improvements. Bugs fixed.
- v0.3(Oct 14, 2023): Way better developer experience. Bugs fixed.
- *v0.4(in progress): E2EE hosting. Charts and Gallery apps.*

Downloads: [GitHub's Releases](https://github.com/MilanSuk/skyalt/releases)



## Compile SkyAlt
SkyAlt is written in Go language. You can install golang from here: https://go.dev/doc/install

Dependencies(sqlite, wazero, websocket, sdl):
<pre><code>go get github.com/mattn/go-sqlite
go get github.com/tetratelabs/wazero
go get github.com/gorilla/websocket
go get github.com/veandco/go-sdl2/sdl
go get github.com/veandco/go-sdl2/ttf
go get github.com/veandco/go-sdl2/gfx
</code></pre>

SkyAlt:
<pre><code>git clone https://github.com/milansuk/skyalt
cd skyalt
go build
./skyalt
</code></pre>



## Create, debug and compile an app
#### Create
- open the main Menu(SkyAlt logo) and under "Developers" click "Create app".
- set the name of the app and the programming language. Click the "Create App" button.
- *note: new app is created in `/apps` folder.*

#### Debug
- in SkyAlt, add your new app under some database(SkyAlt will complain that 'main.wasm' is missing, ignore it).
- open VSCode. From the top menu select "File" -> "Open Folder" and select your app folder. Run the app with the F5 key.
- in SkyAlt, the app will show up with a blue border(that means it's in debug mode).
- *note: you can stop the debugger and run it again as many times as you want. No need to close SkyAlt.*
- *note: When you debug, SkyAlt replaces WASI(WASM interface) with TCP connection. No WASM included.*

#### Compile WASM
- install tinygo compiler with `sudo apt-get install tinygo`.
- run `sh build_wasm` from your app folder.
- `main.wasm` is created in the same folder.
- *note: After this, SkyAlt will stop complaining that 'main.wasm' is missing.*

#### Package
- open the main Menu(SkyAlt logo) and under "Developers" click "Package app".
- select the app. Click the 'Package app' button.
- *note: all app's files are copied into a single `<app_name>.sqlite` file, which you can find in `/apps` folder.*



## Repository
- /apps - application's repos & packages
- /databases - user's databases
- /resources - default fonts, images



## Author
Milan Suk

Email: milan@skyalt.com

Twitter: https://twitter.com/milansuk/

**Sponsor**: https://github.com/sponsors/MilanSuk

*Feel free to follow or contact me with any idea, question or problem.*



## Contributing
Your feedback and code are welcome!

For bug report or question, please use [GitHub's Issues](https://github.com/MilanSuk/skyalt/issues)

SkyAlt is licensed under **Apache v2.0** license. This repository includes 100% of the code.
