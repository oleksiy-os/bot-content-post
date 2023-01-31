## This project created just for practical usage of HUGO & webpages create from telegram bot
Telegram bot for posting new content items to website created by great static site generator [Hugo](https://gohugo.io)

## Steps to make it work
1. [Create website with Hugo](https://gohugo.io/getting-started/quick-start/) or clone from [here](https://github.com/oleksiy-os/hugo-demo) 
2. For created site copy/edit rule from [here](https://github.com/oleksiy-os/hugo-demo/blob/main/archetypes/posts.md). It creates new article from json file inside "data/externalPost/".
3. Rename `config-sample.json` to `config-sample.json`, [create tg bot](https://core.telegram.org/bots/api#authorizing-your-bot) & add token
4. Build go bin, put it & config.json to the hugo website root


## How it works
When GO bin runs from the website root, it waits commands from telegram bot.

If used command `/add`: 
1. bot asks article title and description
2. creates json file with content `data/externalPost/<title>.json`
3. runs cli HUGO command for create new website page `hugo new posts/<title>.md`
4. Hugo has snippet for path `posts/`. Look at folder `data/externalPost/<title>.json`, if exist such file create from it's content the website page