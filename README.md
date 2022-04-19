# rc-postcard
[![Go Report Card](https://goreportcard.com/badge/github.com/rc-postcard/rc-postcard)](https://goreportcard.com/report/github.com/rc-postcard/rc-postcard) [![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0) ![Coverage: O](https://img.shields.io/badge/coverage-200%25-red)

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

## Bonus
- Validate address
- Show better preview with iframe