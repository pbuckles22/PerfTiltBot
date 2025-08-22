#!/usr/bin/env python3
"""
Twitch OAuth Helper Script
Automates the OAuth flow for creating new bot configurations
"""

import requests
import webbrowser
import json
import os
import sys
from urllib.parse import urlencode, parse_qs, urlparse

def create_twitch_app():
    """Guide user through creating a Twitch application"""
    print("=== Twitch Application Setup ===")
    print("1. Go to: https://dev.twitch.tv/console")
    print("2. Click 'Register Your Application'")
    print("3. Fill in the form:")
    print("   - Name: Your bot name (e.g., 'MyAwesomeBot')")
    print("   - OAuth Redirect URLs: http://localhost:3000")
    print("   - Category: Chat Bot")
    print("4. Click 'Create'")
    print("5. Copy the Client ID and Client Secret")
    print()
    
    client_id = input("Enter Client ID: ").strip()
    client_secret = input("Enter Client Secret: ").strip()
    
    return client_id, client_secret

def get_oauth_url(client_id, scopes):
    """Generate OAuth URL"""
    base_url = "https://id.twitch.tv/oauth2/authorize"
    params = {
        'client_id': client_id,
        'redirect_uri': 'http://localhost:3000',
        'response_type': 'code',
        'scope': ' '.join(scopes)
    }
    return f"{base_url}?{urlencode(params)}"

def get_auth_code():
    """Open browser and get authorization code"""
    print("=== Authorization ===")
    print("1. A browser window will open")
    print("2. Log in with your Twitch account")
    print("3. Authorize the application")
    print("4. Copy the 'code' parameter from the URL")
    print()
    
    # Open browser
    webbrowser.open('http://localhost:3000')
    
    print("After authorization, you'll be redirected to a URL like:")
    print("http://localhost:3000?code=abc123...")
    print()
    
    redirect_url = input("Paste the full redirect URL: ").strip()
    
    # Parse the code from URL
    parsed = urlparse(redirect_url)
    params = parse_qs(parsed.query)
    
    if 'code' not in params:
        print("Error: No authorization code found in URL")
        sys.exit(1)
    
    return params['code'][0]

def exchange_code_for_tokens(client_id, client_secret, auth_code):
    """Exchange authorization code for tokens"""
    url = "https://id.twitch.tv/oauth2/token"
    data = {
        'client_id': client_id,
        'client_secret': client_secret,
        'code': auth_code,
        'grant_type': 'authorization_code',
        'redirect_uri': 'http://localhost:3000'
    }
    
    response = requests.post(url, data=data)
    
    if response.status_code != 200:
        print(f"Error: {response.status_code} - {response.text}")
        sys.exit(1)
    
    return response.json()

def refresh_token(client_id, client_secret, refresh_token):
    """Refresh the access token"""
    url = "https://id.twitch.tv/oauth2/token"
    data = {
        'client_id': client_id,
        'client_secret': client_secret,
        'refresh_token': refresh_token,
        'grant_type': 'refresh_token'
    }
    
    response = requests.post(url, data=data)
    
    if response.status_code != 200:
        print(f"Error refreshing token: {response.status_code} - {response.text}")
        return None
    
    return response.json()

def create_bot_config(bot_name, client_id, client_secret, access_token, refresh_token):
    """Create bot configuration file"""
    config = {
        'bot_name': bot_name,
        'oauth': f'oauth:{access_token}',
        'client_id': client_id,
        'client_secret': client_secret,
        'refresh_token': refresh_token
    }
    
    # Ensure configs/bots directory exists
    os.makedirs('configs/bots', exist_ok=True)
    
    # Write config file
    filename = f'configs/bots/{bot_name}_auth_secrets.yaml'
    with open(filename, 'w') as f:
        f.write(f"bot_name: \"{bot_name}\"\n")
        f.write(f"oauth: \"oauth:{access_token}\"\n")
        f.write(f"client_id: \"{client_id}\"\n")
        f.write(f"client_secret: \"{client_secret}\"\n")
        f.write(f"refresh_token: \"{refresh_token}\"\n")
    
    print(f"✅ Bot configuration saved to: {filename}")
    return filename

def create_channel_config(bot_name, channel_name):
    """Create channel configuration file"""
    config = {
        'bot_name': bot_name,
        'channel': channel_name,
        'commands_enabled': True,
        'cooldown_seconds': 30
    }
    
    # Ensure configs/channels directory exists
    os.makedirs('configs/channels', exist_ok=True)
    
    # Write config file
    filename = f'configs/channels/{channel_name}_config_secrets.yaml'
    with open(filename, 'w') as f:
        f.write(f"bot_name: \"{bot_name}\"  # References which bot's auth file to use\n")
        f.write(f"channel: \"{channel_name}\"  # Twitch channel name\n")
        f.write("commands_enabled: true  # Enable/disable bot commands\n")
        f.write("cooldown_seconds: 30  # Cooldown between commands\n")
    
    print(f"✅ Channel configuration saved to: {filename}")
    return filename

def main():
    print("🤖 Twitch Bot OAuth Helper")
    print("This script will help you create a new bot configuration\n")
    
    # Get bot details
    bot_name = input("Enter bot name (e.g., 'MyAwesomeBot'): ").strip()
    channel_name = input("Enter channel name (e.g., 'mychannel'): ").strip()
    
    print(f"\nSetting up bot '{bot_name}' for channel '{channel_name}'...\n")
    
    # Step 1: Create Twitch application
    client_id, client_secret = create_twitch_app()
    
    # Step 2: Get authorization code
    scopes = ['chat:read', 'chat:edit', 'channel:moderate']
    oauth_url = get_oauth_url(client_id, scopes)
    print(f"OAuth URL: {oauth_url}")
    
    auth_code = get_auth_code()
    
    # Step 3: Exchange code for tokens
    print("\nExchanging authorization code for tokens...")
    tokens = exchange_code_for_tokens(client_id, client_secret, auth_code)
    
    access_token = tokens['access_token']
    refresh_token = tokens['refresh_token']
    
    # Step 4: Test token refresh
    print("Testing token refresh...")
    refreshed = refresh_token(client_id, client_secret, refresh_token)
    if refreshed:
        print("✅ Token refresh works!")
    else:
        print("⚠️  Token refresh failed, but continuing...")
    
    # Step 5: Create configuration files
    print("\nCreating configuration files...")
    bot_config = create_bot_config(bot_name, client_id, client_secret, access_token, refresh_token)
    channel_config = create_channel_config(bot_name, channel_name)
    
    print(f"\n🎉 Setup complete!")
    print(f"Bot config: {bot_config}")
    print(f"Channel config: {channel_config}")
    print(f"\nTo deploy to EC2:")
    print(f"1. Copy configs/bots/ and configs/channels/ to EC2")
    print(f"2. Run: ./deploy_ec2_enhanced.sh start {channel_name}")

if __name__ == "__main__":
    main()
