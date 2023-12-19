### Motivation
It's been a while since I've built in Go, and I wanted to add some sort of song recommendation widget for my personal website. This is a super barebones implementation of an API for that using [Go Chi](https://github.com/go-chi/chi) that allows you (and only you, or anyone else with your `API_TOKEN`) to add songs with timestamps. If you want to display them, check out my [personal website.](https://sterp.dev)
### Quickstart
1. Launch a Postgres database with database "song", with schema:
   ```
   id: PK (int)
   title: text
   artist: text
   image_url: text
   song_url: text
   platform: text
   submitted_at: timestamp without timezone 
   ```
2. Register for Spotify API access
3. Modify `.env` file in the root directory and change the following variables to your own:
```
API_TOKEN=ARBITRARY_API_KEY

DB_HOST=your_db_host
DB_USER=your_db_user
DB_PASS=your_db_pass
DB_NAME=your_db_name

SOUNDCLOUD_CLIENT=your_soundcloud_client

SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
```
4. Run `docker-compose up` to build and run the server. You should see something like: `2023/12/19 20:58:13 Server is up and running!`
5. Server is now listening on port 443.

### Todo
- Captcha authenticated song recommendations (from the public to you)
- Soundcloud as a provider, they aren't giving out API access :(
- Potentially forcing all API calls to have an API token, and use a secret token for adding song