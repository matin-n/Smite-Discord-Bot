# Smite-Discord-Bot
Used to lookup Smite player statistics by connecting to HiRez API

## Configuration
1. Replace `devId` and `authKey`
2. Create Discord authentication token for your Discord Bot & replace with `discordAuthToken`
3. Run the bot

## Usage
* `!ranked playername` to lookup ranked statistics
* `!playerStatus playername` to lookup player status (in-game, lobby, etc)
* `!playerID playername` to lookup the player ID

----
## Libraries Used
* [DiscordGo](https://github.com/bwmarrin/discordgo)
* [GJSON](https://github.com/tidwall/gjson)
