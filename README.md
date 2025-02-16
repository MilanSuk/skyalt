<p align="center">
<img src="https://raw.githubusercontent.com/MilanSuk/skyalt/refs/heads/main/screenshots/screenshot_2.png" />
</p>



# Skyalt
A new interface, which offers simplicity, local-first computing and LLM-assistant at the core.

Web: https://www.skyalt.com



## Current state
- **Experimental**
- ~9K LOC for Skyalt and ~10K for widgets code
- Build on Linux, but should support all mainstream operation systems
- Layout and rendering engines are written from scratch
- Use cloud LLM services or go fully local-first with llama.cpp and whisper.cpp.
- More information on https://www.skyalt.com



## Compile Skyalt
Install Go language:
- https://go.dev/doc/install

Install SDL and FFmpeg:
<pre><code>sudo apt-get install libsdl2-dev
sudo apt-get install libsdl2-mixer-dev
sudo apt install ffmpeg
</code></pre>

Install Golang tools:
<pre><code>go install golang.org/x/tools/cmd/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
</code></pre>

Compile Skyalt:
<pre><code>git clone https://github.com/milansuk/skyalt
cd skyalt
go mod tidy
go build
./skyalt
</code></pre>

(optional) Service LLama.cpp(~100MB):
<pre><code>cd services
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp
make
</code></pre>

(optional) Service Whisper.cpp(~30MB):
<pre><code>cd services
git clone https://github.com/ggerganov/whisper.cpp
cd whisper.cpp
make
</code></pre>



## Author
Milan Suk

Email: milan@skyalt.com

X: https://x.com/milansuk/

**Sponsor**: https://github.com/sponsors/milansuk

*Feel free to follow or contact me with any idea, question or problem.*



## Contributing
Your feedback and code are welcome!

For bug report or question, please use [GitHub's Issues](https://github.com/milansuk/skyalt/issues)

Skyalt is licensed under **Apache v2.0** license. This repository includes 100% of the code.
