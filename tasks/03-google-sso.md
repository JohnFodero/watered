# Task 3: Google SSO Integration

## Objective
Implement Google OAuth2 authentication with email whitelist functionality.

## Requirements
- [ ] Set up Google OAuth2 client credentials
- [ ] Implement OAuth2 flow in Go backend
- [ ] Create session management with secure cookies
- [ ] Add email whitelist validation
- [ ] Implement login/logout endpoints
- [ ] Add authentication middleware for protected routes
- [ ] Create user model and session storage
- [ ] Add redirect handling after authentication

## Security Considerations
- [ ] Secure cookie configuration (httpOnly, secure, sameSite)
- [ ] CSRF protection
- [ ] State parameter validation in OAuth flow
- [ ] Token validation and refresh handling
- [ ] Session timeout management

## API Endpoints
- `GET /auth/login` - Redirect to Google OAuth
- `GET /auth/callback` - Handle OAuth callback
- `POST /auth/logout` - Clear session and logout
- `GET /auth/status` - Check authentication status

## Environment Variables
- `GOOGLE_CLIENT_ID` - OAuth2 client ID
- `GOOGLE_CLIENT_SECRET` - OAuth2 client secret
- `SESSION_SECRET` - Session encryption key
- `ALLOWED_EMAILS` - Comma-separated whitelist

## Success Criteria
- Users can authenticate with Google accounts
- Only whitelisted emails can access the app
- Sessions are secure and properly managed
- Authentication state persists across page reloads
- Tests cover authentication flows

## Next Steps
- Task 4: Plant state management API
- Task 5: Admin panel functionality