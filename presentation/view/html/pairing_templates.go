// Package html contains HTML template constants for web pages.
// This package follows Clean Architecture principles by separating presentation markup
// from business logic and HTTP handling.
package html

// PairingSuccessTemplate is the HTML page shown when WhatsApp is already connected.
// It displays a success message with device information.
const PairingSuccessTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Pairing - BotjanWeb</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f0f2f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .success { color: #00a884; font-size: 24px; margin: 20px 0; }
        .device-info { background: #f0f2f5; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .emoji { font-size: 48px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="emoji">✅</div>
        <h1>WhatsApp Already Connected</h1>
        <p class="success">Device is already paired and logged in!</p>
        <div class="device-info">
            <strong>Device ID:</strong> {{.DeviceID}}
        </div>
        <p style="color: #667781; margin-top: 30px;">
            No action needed. The bot is running normally.
        </p>
    </div>
</body>
</html>`

// PairingPageTemplate is the interactive HTML page for WhatsApp QR code pairing.
// It includes QR code display, polling logic, and step-by-step instructions.
const PairingPageTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>WhatsApp Pairing - BotjanWeb</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 20px; background: #f0f2f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .qr-container { margin: 30px 0; padding: 20px; background: #f0f2f5; border-radius: 10px; }
        #qrcode { margin: 20px auto; max-width: 300px; }
        #qrcode img { width: 100%; height: auto; }
        .loading { color: #667781; }
        .instructions { text-align: left; margin: 20px 0; padding: 20px; background: #e7f3ff; border-radius: 5px; }
        .instructions ol { margin: 10px 0; padding-left: 20px; }
        .instructions li { margin: 8px 0; }
        .status { padding: 10px; border-radius: 5px; margin: 15px 0; }
        .status.waiting { background: #fff3cd; color: #856404; }
        .status.success { background: #d4edda; color: #155724; }
        .status.error { background: #f8d7da; color: #721c24; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 WhatsApp Pairing</h1>
        <p style="color: #667781;">Scan the QR code below with WhatsApp to pair your device</p>
        
        <div class="instructions">
            <strong>📱 How to pair:</strong>
            <ol>
                <li>Open WhatsApp on your phone</li>
                <li>Tap <strong>Menu (⋮)</strong> or <strong>Settings</strong></li>
                <li>Tap <strong>Linked Devices</strong></li>
                <li>Tap <strong>Link a Device</strong></li>
                <li>Point your phone at this screen to scan the QR code</li>
            </ol>
        </div>

        <div id="status" class="status waiting">⏳ Waiting for QR code...</div>
        
        <div class="qr-container">
            <div id="qrcode" class="loading">
                <p>Loading QR code...</p>
            </div>
        </div>

        <p style="color: #667781; font-size: 14px;">
            ⚠️ <strong>Security:</strong> Keep this page private. Anyone with this QR can access your WhatsApp.
        </p>
    </div>

    <script>
        let lastQR = '';
        let pollInterval;

        async function fetchQRCode() {
            try {
                const response = await fetch('/pairing/qr?token={{.Token}}');
                const data = await response.json();
                
                if (data.qr && data.qr !== lastQR) {
                    lastQR = data.qr;
                    document.getElementById('qrcode').innerHTML = 
                        '<img src="' + data.qr + '" alt="QR Code">';
                    document.getElementById('status').className = 'status waiting';
                    document.getElementById('status').textContent = '📸 Scan this QR code with WhatsApp';
                } else if (data.status === 'success') {
                    clearInterval(pollInterval);
                    document.getElementById('status').className = 'status success';
                    document.getElementById('status').textContent = '✅ Successfully paired! Reloading...';
                    setTimeout(() => location.reload(), 2000);
                } else if (data.status === 'timeout') {
                    clearInterval(pollInterval);
                    document.getElementById('status').className = 'status error';
                    document.getElementById('status').textContent = '⏱️ QR code expired. Refresh the page to generate a new one.';
                } else if (data.error) {
                    document.getElementById('status').className = 'status error';
                    document.getElementById('status').textContent = '❌ Error: ' + data.error;
                }
            } catch (err) {
                console.error('Failed to fetch QR:', err);
            }
        }

        // Poll for QR code every 2 seconds
        fetchQRCode();
        pollInterval = setInterval(fetchQRCode, 2000);
    </script>
</body>
</html>`
