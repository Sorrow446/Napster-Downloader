# Napster-Downloader
Napster downloader written in Go.
Updated version with lossless support.
![](https://i.imgur.com/O7MSizK.png)
[Pre-compiled binaries](https://github.com/Sorrow446/Napster-Downloader/releases/latest)

## Setup
Input credentials into config file.
Configure any other options if needed.
|Option|Info|
| --- | --- |
|email|Email address.
|password|Password.
|format|Download quality. 1 = 64 Kbps AAC, 2 = 128 Kbps MP3, 3 = 192 Kbps AAC, 4 = 320 Kbps AAC, 5 = FLAC.
|outPath|Where to download to. Path will be made if it doesn't already exist.
|albumTemplate|Album folder naming template. Vars: album, albumArtist, copyright, label, upc, year.
|trackTemplate|Track filename naming template. Vars: album, albumArtist, artist, copyright, isrc, label, title, track, trackPad, trackTotal, upc, year.
|keepCover|true = don't delete covers from album folders.
|useFFmpegEnvVar|true = call FFmpeg from environment variable, false = call from script dir.

## FFmpeg setup
**FFmpeg is needed for ts -> MP4 losslessly for videos. You can skip this if you don't plan to download videos.**

[Windows (gpl)](https://github.com/BtbN/FFmpeg-Builds/releases)    
Linux: `sudo apt install ffmpeg`    
Termux `pkg install ffmpeg`    
If you don't have root in Linux, you can have Napster Downloader look for the binary in the same dir by setting the `useFfmpegEnvVar` option to false.

## Supported Media
|Type|URL example|
| --- | --- |
|Album|`https://web.napster.com/album/alb.60702691`
|Video|`https://play.napster.com/video/659fc58a3eb69030781a4969`

## Usage
Args take priority over the same config file options.

Download two albums:   
`np_dl_x64.exe https://web.napster.com/album/alb.60702691 https://play.napster.com/album/alb.64027739`

Download a single album and from two text files in format 4:   
`np_dl_x64.exe https://web.napster.com/album/alb.60702691 G:\1.txt G:\2.txt -f 4`

```
 _____             _              ____                _           _
|   | |___ ___ ___| |_ ___ ___   |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
| | | | .'| . |_ -|  _| -_|  _|  |  |  | . | | | |   | | . | .'| . | -_|  _|
|_|___|__,|  _|___|_| |___|_|    |____/|___|_____|_|_|_|___|__,|___|___|_|
          |_|

Usage: np_dl_x64.exe [--format FORMAT] [--outpath OUTPATH] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --format FORMAT, -f FORMAT
                         Download format.
                         1 = 64 Kbps AAC, 2 = 128 Kbps MP3, 3 = 192 Kbps AAC, 4 = 320 Kbps AAC, 5 = FLAC. [default: -1]
  --outpath OUTPATH, -o OUTPATH
                         Where to download to. Path will be made if it doesn't already exist.
  --help, -h             display this help and exit
  ```
   
# Disclaimer
- I will not be responsible for how you use Napster Downloader.    
- Napster brand and name is the registered trademark of its respective owner.    
- Napster Downloader has no partnership, sponsorship or endorsement with Napster.
