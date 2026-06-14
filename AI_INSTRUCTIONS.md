# AI Developer Instructions for `smieci-sms`

When adding new environment variables or configuration secrets to the application (e.g., in `config.go`), you **MUST** also update the following deployment files:

1. **`docker-compose.yml`**: Ensure the new variable is passed to the `app` container's `environment` block.
   ```yaml
   environment:
     - NEW_SECRET=${NEW_SECRET}
   ```

2. **`.github/workflows/deploy.yml`**: The deployment script manually generates the `.env` file on the remote server using GitHub Secrets. You **MUST** add an `echo` statement for the new secret in the deployment script.
   ```yaml
   script: |
     cd /app
     ...
     echo "NEW_SECRET=${{ secrets.NEW_SECRET }}" >> .env
   ```

Failure to update both files will result in the production container crashing on startup (which bubbles up as a `502 Bad Gateway` error in Caddy) because the backend fails the `config.Validate()` check.
