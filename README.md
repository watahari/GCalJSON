# GCalJSON

GCalJSON is a lightweight Go API that retrieves events from Google Calendar and returns them in JSON format—specifically structured for Grafana's Business Calendar Plugin. It fetches upcoming events, caches them for a configurable duration, and provides detailed logging and graceful shutdown for robust production use.

## Features

- Retrieves events from Google Calendar via the Calendar API.
- Caches API responses for a configurable duration (default is 5 minutes).
- Outputs JSON data structured for Grafana's **Business Calendar Plugin**.
- Implements detailed error logging and graceful shutdown.
- Containerized with Docker and easily deployable via docker-compose.
- Continuous Integration/Delivery with GitHub Actions for testing and Docker Hub publishing.

## Environment Variables

GCalJSON uses the following environment variables (with the `GCALJSON_` prefix) to configure its behavior:

- **GCALJSON_GOOGLE_CREDENTIALS**  
  A JSON string containing the Google API credentials (Service Account Key).
- **GCALJSON_GOOGLE_CALENDAR_ID**  
  The identifier of the Google Calendar from which events will be fetched.
- **GCALJSON_CACHE_DURATION**  
  Duration for caching the API responses (e.g., `"5m"` for 5 minutes). If not set, the default is 5 minutes.

### .env Sample

Create a `.env` file in your project root with content similar to:

```dotenv
GCALJSON_GOOGLE_CREDENTIALS='{
  "type": "service_account",
  "project_id": "your-project-id",
  "private_key_id": "your-private-key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\nYOUR_PRIVATE_KEY\n-----END PRIVATE KEY-----\n",
  "client_email": "your-service-account-email@your-project-id.iam.gserviceaccount.com",
  "client_id": "your-client-id",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/your-service-account-email@your-project-id.iam.gserviceaccount.com"
}'

GCALJSON_GOOGLE_CALENDAR_ID='your-calendar-id@group.calendar.google.com'
GCALJSON_CACHE_DURATION='5m'
```

## Obtaining Google Credentials and Calendar ID

1. **Google Credentials**:
  * Visit the [Google Cloud Console](https://console.cloud.google.com/).
  * Create a new project or select an existing one.
  * Enable the Google Calendar API for your project.
  * Under APIs & Services > Credentials, create a new Service Account.
  * Download the JSON key file and use its content as the value for `GCALJSON_GOOGLE_CREDENTIALS`.
2. **Google Calendar ID**:
  * Open [Google Calendar](https://calendar.google.com/).
  * Navigate to the settings of the calendar you wish to use.
  * Under the "Integrate calendar" section, locate the Calendar ID.
  * Use this ID as the value for `GCALJSON_GOOGLE_CALENDAR_ID`.
3. **Granting Access to the Calendar**:
  * In order for your service account to access the calendar, you must share the calendar with the service account's email address.
  * Open the calendar settings in Google Calendar.
  * Under "Share with specific people" or "Access permissions", add the service account's email address (found in the `client_email` field of your credentials JSON).
  * Grant the service account "See all event details" (or the necessary permissions required for your application).

## Docker Compose
Below is a sample `docker-compose.yaml` file to run GCalJSON:

```yaml
version: "3"
services:
  gcaljson:
    build: .
    ports:
      - "8080:8080"
    environment:
      - GCALJSON_GOOGLE_CREDENTIALS=${GCALJSON_GOOGLE_CREDENTIALS}
      - GCALJSON_GOOGLE_CALENDAR_ID=${GCALJSON_GOOGLE_CALENDAR_ID}
      - GCALJSON_CACHE_DURATION=${GCALJSON_CACHE_DURATION}
```

Ensure that your environment variables are set either in a `.env` file or directly in your host environment.

## Grafana Configuration
To integrate GCalJSON with Grafana's Business Calendar Plugin, follow these steps:

1. **Install the Plugin**:
  * In Grafana, navigate to Configuration > Plugins.
  * Search for Business Calendar Plugin by Marcus Olsson.
  * Install the plugin (or use the CLI:  
`grafana-cli plugins install marcusolsson-calendar-panel` ).
2. **Add a JSON Data Source**:
  * Go to Configuration > Data Sources in Grafana and click Add data source.
  * Select a JSON API data source plugin (if not already available, install a compatible JSON data source plugin).
  * Set the URL to point to your GCalJSON instance (e.g., `http://<your-server-ip>:8080` ).
3. **Create a Dashboard Panel**:
  * Create a new dashboard panel and choose Business Calendar as the panel type.
  * Configure the panel to use the JSON data source you just added.
  * The JSON response provided by GCalJSON is expected to be an array of events, where each event includes `title`, `start`, and `end` fields. The Business Calendar Plugin will render these events accordingly.

For more details on the plugin, refer to the [Grafana Business Calendar Plugin page](https://grafana.com/grafana/plugins/marcusolsson-calendar-panel/).

## Building and Running
### With Docker Compose
Run the following command in the project directory:

```bash
docker-compose up --build
```

### GitHub Actions
The repository includes a GitHub Actions workflow ( `.github/workflows/ci.yml` ) that:

* Executes tests on pull requests.
* Builds and pushes the Docker image to Docker Hub on pushes to the `main` branch or when a version tag is pushed.

Before using GitHub Actions, set up the following repository secrets:

* `DOCKERHUB_USERNAME`
* `DOCKERHUB_TOKEN`

## License
This project is licensed under the MIT License.
