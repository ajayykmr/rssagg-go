# rssagg-go
An RSS aggregator developed in Golang.

# Sample RSS Feeds Included:
https://www.wagslane.dev/index.xml
https://abcnews.go.com/abcnews/usheadlines
https://wwwcbsnews.com/latest/rss/main
https://timesofindia.indiatimes.com/rssfeedstopstories.cms
https://www.wired.com/feed/rss


## Programming Language 

- Go

## Database

- Postgres SQL
- pgAdmin 4
- [Goose](https://pressly.github.io/goose/) for database migration
- [sqlc](https://sqlc.dev/) for generating type-safe Go code from SQL

## Features

- Users
  - Create a user
  - Get user information
- Feeds
  - Create a feed
  - Get all the feeds
- Feed following
  - Create a feed following
  - Get all the feed following of a user
  - Delete a feed following
- Posts
  - Get updated posts subscribed by a user
