module McQueens_Tea_Cup

go 1.25.0

require (
	github.com/Duke40199/sega-idac-client v0.0.0-00010101000000-000000000000
	github.com/PuerkitoBio/goquery v1.12.0
	github.com/bwmarrin/discordgo v0.29.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	golang.org/x/text v0.37.0
	google.golang.org/grpc v1.68.1
)

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	golang.org/x/crypto v0.49.0 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)

// LOCAL DEV ONLY — do not deploy with this in place. It resolves the dependency
// to a sibling checkout so the repo builds before the first tag is published.
// Go-live: publish/tag sega-idac-client, then:
//   go mod edit -dropreplace=github.com/Duke40199/sega-idac-client
//   go get github.com/Duke40199/sega-idac-client@v0.1.0
// (and set GOPRIVATE=github.com/Duke40199/* for the private module).
replace github.com/Duke40199/sega-idac-client => ../sega-idac-client
