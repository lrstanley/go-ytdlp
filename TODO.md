# TODO

- MustInstall lock of some kind?
- json representation with marshal/unmarshal
- http server endpoint generation
- get go-ytdlp added to <https://github.com/yt-dlp/yt-dlp#embedding-yt-dlp>?
- keep track of supported "%(<format>)s" options?
- PrintToFile (support json -> struct for this?)
- tests for Result and ResultLog
- internal json -> struct map (incl all our cleaned up fields), so we could export all options via graphql for go-ytdlp-web?
- support output multi-writer for custom user-defined writers
- make sure TMPDIR or similar is supported (or overridable) for temp files used when merging with
  something like ffmpeg.
