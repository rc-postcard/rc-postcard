# rc-postcard
[![Go Report Card](https://goreportcard.com/badge/github.com/rc-postcard/rc-postcard)](https://goreportcard.com/report/github.com/rc-postcard/rc-postcard) [![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0) ![Coverage: O](https://img.shields.io/badge/coverage-200%25-red)

## How to use website
- Find an image you like and resize to height:width ratio of 4.25:6.25. (Feature request: auto resize)
- I recommend using https://redketchup.io/image-resizer. Use Aspect Ratio "Custom Aspect Ratio" with horizontal ratio 6.25 and vertical ratio 4.25. Then scroll down and "save image"
- Sign in at https://rc-postcard.fly.dev and then create your account with a valid address you'd like to receive mail at. (Sorry US only for now, our top engineers are working on it).
- Upload your resized photo, enter some text, and preview!
- If you like your postcard, select a recipient (from recursers who have signed up) and send!


## Build and Run
Create an OAuth application at [https://www.recurse.com/settings/apps](https://www.recurse.com/settings/apps) with proper redirect URI (http://localhost:8080/auth for local run).
**Make sure to set your app's ID, Secret, and Redirect URI in your environmental variables** (see [.env.example](.env.example)). You can optionally set your own redis host and password.

Create an account on [lob.com](https://lob.com) and set your API keys in your environment variables (see [.env.example](.env.example)).

```shell
# Load your environmental variables after setting them
# Note: we recommend copying the blank .env.example into a .env file and setting your environmental variables there.
🎨 source .env

# Run rc-postcard app
🎨 make run
```
🎉 rc-postcard should now be running at [http://localhost:8080](http://localhost:8080)

## Other tools
TODO: make tools to make postcard management easier

## TODO before launch
- disclaimer
- credit users?

## Bonus
- Resize image
- Validate address
- Show better preview with iframe?
- RC logo