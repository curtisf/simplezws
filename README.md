# SimpleZWS

SimpleZWS is a *simple* image uploader and URL "hider" for personal use. Many large chat platforms will clean ZWS characters from a URL, but it works on Discord, where I am using this.

## Installation

Clone the repo

```bash
git clone https://github.com/caf203/simplezws && cd simplezws
```

Populate your config (with your editor of choice)

```bash
mv .config.example .config
```

Create the image directory

```bash
mkdir images
```

## Usage

Use your tool of choice to POST to the base URL you provided in your config. This is how I do:

```bash
maim -s --format=png /dev/stdout | curl -H "Authorization: yourauthstring" -F 'data=@-' https://curtisf.dev/u | xclip -selection clipboard
```

The response returned will be your base URL + the ZWS characters following it.

## License
[MIT](https://choosealicense.com/licenses/mit/)