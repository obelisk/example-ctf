
# CTF Platform

A Capture The Flag competition featuring regular challenges and an advanced exam section with token-based mechanics.

## 🏆 CTF Rules & Game Mechanics

### Overview
This CTF consists of two main challenge types:
- **Regular Challenges**: Earn tokens and points by solving various categories of challenges
- **Exam Challenges**: Advanced challenges that require tokens to attempt

### Token System
- **Earning Tokens**: Complete regular challenges to earn 1 token each
- **Burning Tokens**: Each exam challenge submission costs 1 token (whether correct or incorrect)
- **Token Balance**: Track your available and burned tokens

### Challenge Categories
- **Regular Challenges**: Crypto, web, forensics, and other categories
- **Exam Challenges**: Advanced challenges with sequential progression

### Exam Challenge Rules
- **Sequential Access**: You can only access exam challenges in order
- **Token Cost**: Each submission (correct or incorrect) costs 1 token
- **Progression**: Complete one exam challenge to unlock the next
- **No Points**: Exam challenges don't award points, only tokens

### Scoring System
- **Points**: Earned from regular challenges (varies by challenge difficulty) - helps gauge challenge difficulty
- **Tokens**: Used for exam challenges
- **Ranking**: Determined by the following criteria in order:
  1. **Most exam challenges completed** - Primary ranking factor
  2. **Fastest completion time** - Among users with same exam progress
  3. **Most points** - Tiebreaker for users with same exam progress and completion time

### Additional Features
- **User Aliases**: Set custom aliases for leaderboard display
- **History Logging**: All attempts and completions are logged
- **Slack Integration**: Challenge completions posted to Slack
- **File Downloads**: Some challenges include downloadable assets

## 🚀 Setup Instructions

### Prerequisites
- Docker and Docker Compose
- AWS CLI configured (for asset storage)
- SSL certificates for HTTPS

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd folder
   ```

2. **Configure environment variables**
   ```bash
   cd web-server/backend
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Set up SSL certificates**
   ```bash
   cd web-server/nginx_conf/certs
   # Place your SSL certificate files here
   # - cert.pem
   # - key.pem
   ```

4. **Start the services**
   ```bash
   cd web-server
   docker-compose up -d
   ```

5. **Initialize the database**
   ```bash
   cd web-server/sql
   ./migrate-docker.sh ./migration.sql
   ```

6. **Access the application**
   - Frontend: https://your-domain.com
   - API: https://your-domain.com/api
   - Health check: https://your-domain.com/health

### Configuration

#### Backend Configuration (`web-server/backend/config/config.yaml`)
```yaml
http:
  port: 8080
  rateLimit:
    enabled: true
    requestsPerSec: 4.0
    burstSize: 20

auth:
  expectedVerifiedAccessInstanceArn: "your-va-instance-arn"
  expectedIssuer: "https://your-sso-provider.com"
  awsRegion: "us-east-2"

database:
  hostname: "postgres"
  port: "50052"
  user: "33ccdb2917"
  database: "ctfservice"

awsConfig:
  bucketName: "your-s3-bucket"

slack:
  leaderboardInterval: "30m"
```

#### Environment Variables (`web-server/backend/.env`)
```bash
# Database
POSTGRES_PASSWORD=your-db-password

# AWS
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=us-east-2

# Slack
SLACK_WEBHOOK_URL=your-slack-webhook
SLACK_BOT_TOKEN=your-slack-bot-token

# Auth
VA_INSTANCE_ARN=your-verified-access-instance-arn
```

### Architecture

#### Components
- **Backend**: Go web server with authentication and challenge management
- **Database**: PostgreSQL for user data and challenge tracking
- **Frontend**: Simple HTML/JS interface
- **Nginx**: Reverse proxy with SSL termination
- **AWS S3**: Asset storage for challenge files

#### Services
- **PostgreSQL**: Database server (port 50052)
- **Backend**: Go API server (port 8080)
- **Nginx**: Web server with SSL (port 443)

### API Endpoints

#### Authentication
- `POST /login` - User login
- `POST /logout` - User logout

#### User Management
- `GET /user/` - Get user profile and progress
- `POST /user/alias` - Set user alias
- `DELETE /user/alias` - Remove user alias

#### Regular Challenges
- `GET /challenges` - List all regular challenges
- `GET /challenges/{id}` - Get specific challenge details
- `POST /challenges/submit` - Submit challenge flag

#### Exam Challenges
- `GET /exam` - List available exam challenges
- `GET /exam/{id}` - Get specific exam challenge
- `POST /exam/submit` - Submit exam challenge flag

### Database Schema

#### Core Tables
- `challenges` - Challenge definitions and metadata
- `flags` - Challenge flags and validation handlers
- `users` - User tokens, points, and exam progress
- `user_challenges_completed` - Challenge completion tracking
- `user_history_log` - User activity logging
- `user_aliases` - User alias management

#### Key Fields
- `tokens_available` - Current token balance
- `tokens_burned` - Total tokens spent on exam challenges
- `points_achieved` - Total points from regular challenges
- `exam_challenges_solved` - Number of completed exam challenges

### Development

#### Adding Challenges
1. Insert challenge data into `challenges` table
2. Add flag and validation handler to `flags` table
3. Upload challenge files to S3 bucket
4. Update challenge metadata with file asset references

#### Custom Validation Handlers
The system supports custom validation logic for flags:
- `exact` - Exact string match
- `case_insensitive` - Case-insensitive match
- `regex` - Regular expression match
- Custom handlers can be implemented in the backend

### Monitoring

#### Logging
- All challenge attempts are logged to `user_history_log`
- Slack notifications for challenge completions
- Rate limiting and security events logged

### Security Features

#### Rate Limiting
- 4 requests per second per client
- Burst allowance of 20 requests
- Maximum 512 concurrent clients

#### Authentication
- AWS Verified Access integration
- SSO provider authentication
- Session management

#### Input Validation
- Flag sanitization and validation
- Challenge ID range validation (1-128)
- SQL injection prevention

### Troubleshooting

#### Logs
```bash
# View all container logs
docker-compose logs

# View specific service logs
docker-compose logs backend
docker-compose logs postgres
docker-compose logs nginx
```
