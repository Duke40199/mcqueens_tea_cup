# McQueen's Tea Cup
A multi-purpose bot for Vietnam's Initial D The Arcade server. Main features include:
- Spamming random things.
- Initial D The Arcade features:
  - Get Time Attack (TA) by track with selective variants / players / countries / cars.
  - Compare TA results between players.
  - Cron-job to crawl a list of most used cars / active players in Online Battles (OB).

## Technical Info:
- Programming languages / Frameworks: Go, Discord Go SDK, PostgreSQL
- Database hosting: Supabase
- CI / CD: Github Actions / Railway
## Project Structure
- Design Pattern: Clean Architecture / Hexagonal Architecture
- File structure:
```mcqueens_tea_cup/
mcqueens_tea_cup/
│   .env
│   .env.example
│   .gitignore
│   go.mod
│   go.sum
│   main.go
│   README.md    
├───cmd
│   └───mckween
│           main.go
│           
└───internal
    ├───adapter
    │   ├───client
    │   │   └───sega_idac
    │   │           client.go
    │   │           
    │   ├───database
    │   │       postgres.go
    │   │       repository.go
    │   │       
    │   ├───discord
    │   │   │   discord_session.go
    │   │   │   handler.go
    │   │   │   handler_cfs.go
    │   │   │   handler_helper.go
    │   │   │   handler_idac.go
    │   │   │   handler_idac_cars.go
    │   │   │   handler_idac_ob.go
    │   │   │   handler_idac_player.go
    │   │   │   handler_shitpost.go
    │   │   │   router.go
    │   │   │   router_definition.go
    │   │   │   
    │   │   └───resource
    │   │           desuwa.gif
    │   │           meomeo.json
    │   │           miemebell.json
    │   │           nuhuh.gif
    │   │           unauthorized.gif
    │   │           
    │   └───repository
    │           alias.go
    │           alias_repo.go
    │           all_net_store_locations_repo.go
    │           area_repo.go
    │           car_repo.go
    │           cfs_state_repo.go
    │           ob_ranking_cfg_repo.go
    │           ranking_cfg_repo.go
    │           ta_time_metadata_repo.go
    │           
    ├───config
    │       config.go
    │       config.json
    │       
    └───domain
        ├───entity
        │       alias.go
        │       cfs_state.go
        │       config.go
        │       idac.go
        │       interface.go
        │       models.go
        │       ob_ranking_cfg.go
        │       ranking_cfg.go
        │       store_location.go
        │       
        ├───port
        │       client.go
        │       service.go
        │       
        └───service
                active_player_sync_service.go
                car_sync_service.go
                discord_writer_service.go
                meta_logic_service.go
                meta_sync_service.go
                store_location_service.go
```
