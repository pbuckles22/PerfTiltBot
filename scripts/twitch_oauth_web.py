#!/usr/bin/env python3
"""
Twitch OAuth Web Helper
Web-based interface for creating new bot configurations
Runs on EC2 and accessible via browser
"""

import requests
import json
import os
import sys
import threading
import time
from urllib.parse import urlencode, parse_qs, urlparse
from flask import Flask, render_template_string, request, redirect, url_for, flash, jsonify
import secrets

app = Flask(__name__)
app.secret_key = secrets.token_hex(16)

# Store OAuth state
oauth_state = {}

# HTML template for the web interface
HTML_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Twitch Bot OAuth Helper</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #9146ff;
            text-align: center;
            margin-bottom: 30px;
        }
        .step {
            background: #f8f9fa;
            padding: 20px;
            margin: 20px 0;
            border-radius: 8px;
            border-left: 4px solid #9146ff;
        }
        .step h3 {
            margin-top: 0;
            color: #333;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #555;
        }
        input[type="text"], input[type="password"] {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
            box-sizing: border-box;
        }
        .btn {
            background: #9146ff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            text-decoration: none;
            display: inline-block;
            margin: 5px;
        }
        .btn:hover {
            background: #7c3aed;
        }
        .btn-secondary {
            background: #6c757d;
        }
        .btn-secondary:hover {
            background: #5a6268;
        }
        .success {
            background: #d4edda;
            color: #155724;
            padding: 15px;
            border-radius: 4px;
            margin: 10px 0;
        }
        .error {
            background: #f8d7da;
            color: #721c24;
            padding: 15px;
            border-radius: 4px;
            margin: 10px 0;
        }
        .info {
            background: #d1ecf1;
            color: #0c5460;
            padding: 15px;
            border-radius: 4px;
            margin: 10px 0;
        }
        .oauth-url {
            background: #e9ecef;
            padding: 10px;
            border-radius: 4px;
            font-family: monospace;
            word-break: break-all;
            margin: 10px 0;
        }
        .hidden {
            display: none;
        }
        .loading {
            text-align: center;
            padding: 20px;
        }
        .spinner {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #9146ff;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 10px;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 Twitch Bot OAuth Helper</h1>
        
        {% with messages = get_flashed_messages(with_categories=true) %}
            {% if messages %}
                {% for category, message in messages %}
                    <div class="{{ category }}">{{ message }}</div>
                {% endfor %}
            {% endif %}
        {% endwith %}

        {% if step == 'setup' %}
        <div class="step">
            <h3>Step 1: Bot Information</h3>
            <form method="POST" action="{{ url_for('setup_bot') }}">
                <div class="form-group">
                    <label for="bot_name">Bot Name:</label>
                    <input type="text" id="bot_name" name="bot_name" placeholder="e.g., MyAwesomeBot" required>
                </div>
                <div class="form-group">
                    <label for="channel_name">Channel Name:</label>
                    <input type="text" id="channel_name" name="channel_name" placeholder="e.g., mychannel" required>
                </div>
                <button type="submit" class="btn">Continue to Twitch App Setup</button>
            </form>
        </div>

        {% elif step == 'twitch_app' %}
        <div class="step">
            <h3>Step 2: Create Twitch Application</h3>
            <div class="info">
                <strong>Follow these steps:</strong>
                <ol>
                    <li>Go to <a href="https://dev.twitch.tv/console" target="_blank">https://dev.twitch.tv/console</a></li>
                    <li>Click "Register Your Application"</li>
                    <li>Fill in the form:
                        <ul>
                            <li><strong>Name:</strong> {{ bot_name }}</li>
                            <li><strong>OAuth Redirect URLs:</strong> http://{{ request.host }}/oauth/callback</li>
                            <li><strong>Category:</strong> Chat Bot</li>
                        </ul>
                    </li>
                    <li>Click "Create"</li>
                    <li>Copy the Client ID and Client Secret</li>
                </ol>
            </div>
            <form method="POST" action="{{ url_for('oauth_start') }}">
                <input type="hidden" name="bot_name" value="{{ bot_name }}">
                <input type="hidden" name="channel_name" value="{{ channel_name }}">
                <div class="form-group">
                    <label for="client_id">Client ID:</label>
                    <input type="text" id="client_id" name="client_id" required>
                </div>
                <div class="form-group">
                    <label for="client_secret">Client Secret:</label>
                    <input type="password" id="client_secret" name="client_secret" required>
                </div>
                <button type="submit" class="btn">Start OAuth Process</button>
                <a href="{{ url_for('index') }}" class="btn btn-secondary">Back</a>
            </form>
        </div>

        {% elif step == 'oauth' %}
        <div class="step">
            <h3>Step 3: OAuth Authorization</h3>
            <div class="info">
                Click the button below to authorize your bot with Twitch. You'll be redirected to Twitch to log in and authorize the application.
            </div>
            <div class="oauth-url">
                {{ oauth_url }}
            </div>
            <a href="{{ oauth_url }}" class="btn" target="_blank">Authorize with Twitch</a>
            <a href="{{ url_for('index') }}" class="btn btn-secondary">Back</a>
        </div>

        {% elif step == 'callback' %}
        <div class="step">
            <h3>Step 4: Processing Authorization</h3>
            <div class="loading">
                <div class="spinner"></div>
                <p>Processing your authorization...</p>
            </div>
        </div>

        {% elif step == 'success' %}
        <div class="step">
            <h3>🎉 Setup Complete!</h3>
            <div class="success">
                <strong>Bot configuration created successfully!</strong>
            </div>
            <div class="info">
                <strong>Files created:</strong>
                <ul>
                    <li>Bot config: {{ bot_config }}</li>
                    <li>Channel config: {{ channel_config }}</li>
                </ul>
            </div>
            <div class="info">
                <strong>Next steps:</strong>
                <ol>
                    <li>Copy the config files to your EC2 instance:</li>
                    <div class="oauth-url">
                        scp -i TwitchViewerGames.pem configs/bots/ ec2-user@{{ ec2_ip }}:~/PerfTiltBot/configs/<br>
                        scp -i TwitchViewerGames.pem configs/channels/ ec2-user@{{ ec2_ip }}:~/PerfTiltBot/configs/
                    </div>
                    <li>SSH to your EC2 instance and start the bot:</li>
                    <div class="oauth-url">
                        ssh -i TwitchViewerGames.pem ec2-user@{{ ec2_ip }}<br>
                        cd ~/PerfTiltBot<br>
                        ./deploy_ec2_enhanced.sh start {{ channel_name }}
                    </div>
                </ol>
            </div>
            <a href="{{ url_for('index') }}" class="btn">Create Another Bot</a>
        </div>
        {% endif %}
    </div>

    <script>
        // Auto-refresh for callback processing
        {% if step == 'callback' %}
        setTimeout(function() {
            window.location.href = "{{ url_for('process_callback') }}";
        }, 2000);
        {% endif %}
    </script>
</body>
</html>
"""

def get_oauth_url(client_id, scopes):
    """Generate OAuth URL"""
    base_url = "https://id.twitch.tv/oauth2/authorize"
    params = {
        'client_id': client_id,
        'redirect_uri': f'http://{request.host}/oauth/callback',
        'response_type': 'code',
        'scope': ' '.join(scopes)
    }
    return f"{base_url}?{urlencode(params)}"

def exchange_code_for_tokens(client_id, client_secret, auth_code):
    """Exchange authorization code for tokens"""
    url = "https://id.twitch.tv/oauth2/token"
    data = {
        'client_id': client_id,
        'client_secret': client_secret,
        'code': auth_code,
        'grant_type': 'authorization_code',
        'redirect_uri': f'http://{request.host}/oauth/callback'
    }
    
    response = requests.post(url, data=data)
    
    if response.status_code != 200:
        raise Exception(f"Token exchange failed: {response.status_code} - {response.text}")
    
    return response.json()

def create_bot_config(bot_name, client_id, client_secret, access_token, refresh_token):
    """Create bot configuration file"""
    # Ensure configs/bots directory exists
    os.makedirs('configs/bots', exist_ok=True)
    
    # Write config file
    filename = f'configs/bots/{bot_name}_auth_secrets.yaml'
    with open(filename, 'w') as f:
        f.write(f'bot_name: "{bot_name}"\n')
        f.write(f'oauth: "oauth:{access_token}"\n')
        f.write(f'client_id: "{client_id}"\n')
        f.write(f'client_secret: "{client_secret}"\n')
        f.write(f'refresh_token: "{refresh_token}"\n')
    
    return filename

def create_channel_config(bot_name, channel_name):
    """Create channel configuration file"""
    # Ensure configs/channels directory exists
    os.makedirs('configs/channels', exist_ok=True)
    
    # Write config file
    filename = f'configs/channels/{channel_name}_config_secrets.yaml'
    with open(filename, 'w') as f:
        f.write(f'bot_name: "{bot_name}"  # References which bot\'s auth file to use\n')
        f.write(f'channel: "{channel_name}"  # Twitch channel name\n')
        f.write('commands_enabled: true  # Enable/disable bot commands\n')
        f.write('cooldown_seconds: 30  # Cooldown between commands\n')
    
    return filename

@app.route('/')
def index():
    return render_template_string(HTML_TEMPLATE, step='setup')

@app.route('/setup', methods=['POST'])
def setup_bot():
    bot_name = request.form.get('bot_name')
    channel_name = request.form.get('channel_name')
    
    if not bot_name or not channel_name:
        flash('Please provide both bot name and channel name', 'error')
        return redirect(url_for('index'))
    
    oauth_state['bot_name'] = bot_name
    oauth_state['channel_name'] = channel_name
    
    return render_template_string(HTML_TEMPLATE, step='twitch_app', bot_name=bot_name, channel_name=channel_name)

@app.route('/oauth/start', methods=['POST'])
def oauth_start():
    client_id = request.form.get('client_id')
    client_secret = request.form.get('client_secret')
    
    if not client_id or not client_secret:
        flash('Please provide both Client ID and Client Secret', 'error')
        return redirect(url_for('index'))
    
    oauth_state['client_id'] = client_id
    oauth_state['client_secret'] = client_secret
    
    scopes = ['chat:read', 'chat:edit', 'channel:moderate']
    oauth_url = get_oauth_url(client_id, scopes)
    
    return render_template_string(HTML_TEMPLATE, step='oauth', oauth_url=oauth_url)

@app.route('/oauth/callback')
def oauth_callback():
    return render_template_string(HTML_TEMPLATE, step='callback')

@app.route('/oauth/process')
def process_callback():
    try:
        # Get authorization code from URL parameters
        auth_code = request.args.get('code')
        if not auth_code:
            flash('No authorization code received', 'error')
            return redirect(url_for('index'))
        
        # Exchange code for tokens
        tokens = exchange_code_for_tokens(
            oauth_state['client_id'],
            oauth_state['client_secret'],
            auth_code
        )
        
        access_token = tokens['access_token']
        refresh_token = tokens['refresh_token']
        
        # Create configuration files
        bot_config = create_bot_config(
            oauth_state['bot_name'],
            oauth_state['client_id'],
            oauth_state['client_secret'],
            access_token,
            refresh_token
        )
        
        channel_config = create_channel_config(
            oauth_state['bot_name'],
            oauth_state['channel_name']
        )
        
        # Get EC2 IP (you'll need to set this)
        ec2_ip = os.environ.get('EC2_IP', '54.190.186.177')
        
        # Clear OAuth state
        oauth_state.clear()
        
        return render_template_string(
            HTML_TEMPLATE,
            step='success',
            bot_config=bot_config,
            channel_config=channel_config,
            channel_name=oauth_state.get('channel_name', ''),
            ec2_ip=ec2_ip
        )
        
    except Exception as e:
        flash(f'Error processing OAuth: {str(e)}', 'error')
        return redirect(url_for('index'))

if __name__ == '__main__':
    import argparse
    
    parser = argparse.ArgumentParser(description='Twitch OAuth Web Helper')
    parser.add_argument('--host', default='0.0.0.0', help='Host to bind to (default: 0.0.0.0)')
    parser.add_argument('--port', type=int, default=3000, help='Port to bind to (default: 3000)')
    parser.add_argument('--debug', action='store_true', help='Enable debug mode')
    
    args = parser.parse_args()
    
    print(f"🤖 Starting Twitch OAuth Web Helper...")
    print(f"🌐 Web interface will be available at: http://{args.host}:{args.port}")
    print(f"📝 Make sure your EC2 security group allows inbound traffic on port {args.port}")
    print(f"🔒 Only accessible from your IP addresses defined in the security group")
    print(f"⏹️  Press Ctrl+C to stop the server")
    print()
    
    app.run(host=args.host, port=args.port, debug=args.debug)
