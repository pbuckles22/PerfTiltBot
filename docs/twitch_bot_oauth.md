# Twitch Bot OAuth: How to Obtain Access and Refresh Tokens

This guide explains how to securely obtain a Twitch OAuth **access token** and **refresh token** for your bot using the Twitch Developer Console and the official `twitch-cli` tool.

---

## **Step 1: Register Your Application on the Twitch Developer Console**

1. Go to the [Twitch Developer Console](https://dev.twitch.tv/console/apps).
2. Click **"Register Your Application"**.
3. Fill in:
   - **Name:** (e.g., `MyBot`)
   - **OAuth Redirect URL:** (e.g., `http://localhost:3000`)
   - **Category:** Chat Bot
4. Click **Create**.
5. On the app page, copy your **Client ID** and **Client Secret**.

---

## **Step 2: Configure `twitch-cli` with Your Credentials**

1. Install `twitch-cli` ([download here](https://github.com/twitchdev/twitch-cli/releases)).
2. Open a terminal and run:
   ```sh
   twitch configure
   ```
3. Enter your **Client ID** and **Client Secret** when prompted.

---

## **Step 3: Get an Access Token and Refresh Token**

1. Run the following command (replace scopes as needed):
   ```sh
   twitch token -u --scopes "chat:read chat:edit"
   ```
   - The `-u` or `--user` flag triggers the user flow, which returns both an access token and a refresh token.
   - This will open a browser window for you to log in and authorize the app.

2. After authorizing, the CLI will output:
   ```
   Access Token: xxxxx
   Refresh Token: xxxxx
   Expires In: 14400
   Scopes: chat:read chat:edit
   ```

3. **Copy both tokens** and use them in your bot configuration (e.g., `bot_auth_secrets.yaml`).

---

## **Step 4: (Optional) Refresh the Token Later**

When your access token expires, you can refresh it using:
```sh
 twitch token --refresh --scopes "chat:read chat:edit"
```
- This will use your stored refresh token and credentials to get a new access token (and possibly a new refresh token).

---

## **Security Notes**
- Never share your client secret or refresh token publicly.
- Store them securely (e.g., in your `bot_auth_secrets.yaml`).
- Do **not** commit secrets to version control.

---

## **Troubleshooting**
- If you only get an access token (no refresh token), make sure you used the `-u` or `--user` flag.
- If you have issues with scopes, double-check the required scopes for your bot's features.
- If you need to manage multiple bots/apps, use `twitch configure --profile <name>` and `twitch token --profile <name> ...`.

---

## **References**
- [Twitch Developer Console](https://dev.twitch.tv/console/apps)
- [twitch-cli GitHub](https://github.com/twitchdev/twitch-cli)
- [twitch-cli token docs](https://github.com/twitchdev/twitch-cli/blob/main/docs/token.md) 