package db

import (
	"context"
	"strconv"
)

// GetMemberGuilds returns a list of guilds that the member is in.
func GetMemberGuilds(ctx context.Context, hostname string, memberID uint64) (guilds []string, err error) {
	query := "SELECT hostname, guild_id FROM guild_members WHERE user_hostname = $1 AND user_id = $2"
	rows, err := dbConn().Query(ctx, query, hostname, memberID)
	if err != nil {
		return
	}
	guilds = []string{}
	defer rows.Close()
	for rows.Next() {
		var guildID uint64
		err = rows.Scan(&hostname, &guildID)
		if err != nil {
			return
		}
		guilds = append(guilds, strconv.FormatUint(guildID, 10)+"@"+hostname)
	}
	return
}

// GetGuildScanners is used to get a guild scanner for every guild ID specified.
func GetGuildScanners(ctx context.Context, guildIds []uint64) (guildScanners []func(any) error, err error) {
	// TODO
	return
}
