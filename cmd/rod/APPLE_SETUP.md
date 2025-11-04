# Sign in with Apple Setup

Guide to configure "Sign in with Apple" OAuth for PocketBase.

## Apple vs Google OAuth

**Apple OAuth** (this guide):
- Requires paid Apple Developer Account ($99/year)
- Uses JWT-based authentication with private key (.p8 file)
- User data (email/name) only provided on FIRST login
- Privacy-focused: users can hide email with privaterelay.appleid.com
- Users must have 2FA enabled on Apple ID

**Google OAuth** (see [GCP_SETUP.md](GCP_SETUP.md)):
- Free Google Cloud account
- Uses client ID + client secret
- User data provided on every login
- Full access to Google Calendar API

## Prerequisites

- **Apple Developer Account** (paid membership - $99/year)
- Apple ID with two-factor authentication enabled
- 15-20 minutes

## Steps

### 1. Join Apple Developer Program

1. Go to https://developer.apple.com/programs/
2. Click **Enroll**
3. Sign in with your Apple ID
4. Complete enrollment (requires payment)
5. Wait for approval (usually 24-48 hours)

### 2. Create App ID

1. Go to https://developer.apple.com/account/resources/identifiers/list
2. Click the **+** button
3. Select **App IDs** → **Continue**
4. Select **App** → **Continue**
5. Fill in:
   - Description: `Wellknown Calendar`
   - Bundle ID: `com.wellknown.calendar` (or your domain reversed)
6. Scroll down to **Capabilities**
7. Check **Sign In with Apple** → **Edit**
8. Choose **Enable as a primary App ID**
9. Click **Save** → **Continue** → **Register**

### 3. Create Services ID (Client ID)

1. Go to https://developer.apple.com/account/resources/identifiers/list/serviceId
2. Click the **+** button
3. Select **Services IDs** → **Continue**
4. Fill in:
   - Description: `Wellknown Calendar Web`
   - Identifier: `com.wellknown.calendar.web` (this becomes your Client ID)
5. Check **Sign In with Apple**
6. Click **Configure** next to Sign In with Apple
7. Primary App ID: Select the App ID created in step 2
8. Domains and Subdomains:
   - For development: `localhost`
   - For production: `yourdomain.com`
9. Return URLs:
   - For development: `http://localhost:8090/auth/apple/callback`
   - For production: `https://yourdomain.com/auth/apple/callback`
10. Click **Save** → **Continue** → **Register**

### 4. Create Private Key

1. Go to https://developer.apple.com/account/resources/authkeys/list
2. Click the **+** button
3. Key Name: `Wellknown Calendar Auth Key`
4. Check **Sign In with Apple**
5. Click **Configure** → Select your App ID → **Save**
6. Click **Continue** → **Register**
7. Click **Download** to get the `.p8` file
   - **IMPORTANT**: This is your ONLY chance to download this file!
   - Save it as `AuthKey_XXXXXXXXXX.p8` (where X is your Key ID)
8. Note the **Key ID** (10 characters, e.g., `ABC123DEF4`)

### 5. Get Team ID

1. Go to https://developer.apple.com/account
2. At the top right, you'll see your Team ID (10 characters, e.g., `XYZ789ABC1`)
3. Copy this value

### 6. Configure PocketBase

Move your private key file:
```bash
# Create directory for Apple credentials
mkdir -p pb/base/secrets

# Move the downloaded .p8 file
mv ~/Downloads/AuthKey_ABC123DEF4.p8 pb/base/secrets/
chmod 600 pb/base/secrets/AuthKey_ABC123DEF4.p8
```

Edit `pb/base/.env` and add:
```bash
# Apple OAuth Configuration
APPLE_TEAM_ID=XYZ789ABC1
APPLE_CLIENT_ID=com.wellknown.calendar.web
APPLE_KEY_ID=ABC123DEF4
APPLE_PRIVATE_KEY_PATH=./secrets/AuthKey_ABC123DEF4.p8
APPLE_REDIRECT_URL=http://localhost:8090/auth/apple/callback
```

Add to `.gitignore`:
```bash
pb/base/secrets/
```

### 7. Test OAuth Flow

```bash
make pb-server
```

Then visit: http://localhost:8090/
- Click **Sign in with Apple**
- Authenticate with Apple ID
- First login: Apple asks for email/name sharing preference
- Subsequent logins: Automatic (email/name not provided again)

## Important Notes

### Email Privacy
- Users can choose to hide their real email
- Hidden emails use format: `abc123@privaterelay.appleid.com`
- Apple forwards emails to user's real address
- You must handle these relay emails correctly

### User Data Caching
- **CRITICAL**: Email and name are ONLY provided on first login
- You MUST store this data in your database immediately
- Subsequent logins only provide the user ID (`sub` claim)
- PocketBase implementation must handle this correctly

### Key Rotation
- Private keys don't expire automatically
- Apple recommends rotating keys periodically
- You can have multiple active keys
- Update `APPLE_KEY_ID` and `APPLE_PRIVATE_KEY_PATH` when rotating

## Troubleshooting

### Error: "invalid_client"
- Check `APPLE_CLIENT_ID` matches your Services ID
- Verify Services ID has "Sign In with Apple" enabled
- Ensure Services ID is configured with correct redirect URL

### Error: "invalid_request"
- Check redirect URL matches exactly (http vs https, trailing slash)
- Verify domain is added to Services ID configuration

### Error: "Invalid JWT"
- Check `APPLE_TEAM_ID` is correct
- Verify `APPLE_KEY_ID` matches your key
- Ensure `.p8` file path is correct and readable
- Check file permissions: `chmod 600 AuthKey_*.p8`

### Error: "The operation couldn't be completed"
- Make sure user has 2FA enabled on Apple ID
- Try signing out of iCloud and back in

### User email/name is null
- This is normal on subsequent logins!
- Check your database - data should be from first login
- If truly missing, user needs to revoke app access and sign in again:
  - iOS: Settings → Apple ID → Password & Security → Apps Using Apple ID
  - macOS: System Preferences → Apple ID → Password & Security

## Production Deployment

For production:
1. Update Services ID configuration:
   - Add your production domain
   - Add production redirect URL
2. Update `APPLE_REDIRECT_URL` in `.env`
3. Consider using environment variables instead of `.env` file
4. Store `.p8` file securely (vault, secrets manager)
5. Never commit `.p8` file to git
6. Submit your app for App Store review if building iOS/macOS app

## Resources

- Apple Developer Portal: https://developer.apple.com/account
- Sign in with Apple Docs: https://developer.apple.com/sign-in-with-apple/
- OpenID Configuration: https://appleid.apple.com/.well-known/openid-configuration
- Token Endpoint: https://appleid.apple.com/auth/token
- Authorization Endpoint: https://appleid.apple.com/auth/authorize
