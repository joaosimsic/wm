module wm

go 1.26.5

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/BurntSushi/xgb v0.0.0-20210121224620-deaf085860bc
)

replace github.com/BurntSushi/xgb => ./third_party/xgb
