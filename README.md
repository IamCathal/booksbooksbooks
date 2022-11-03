# booksbooksbooks

I use GoodReads for my book reccomendations and I buy my books mostly from TheBookshop.ie. I made this project to be able to automatically check if books on my "shopping list" are available on TheBookshop. Still in active development but hoping to make a 1.0 release before the end of November 2022.

|      |  |
| ----------- | ----------- |
| ![](https://i.imgur.com/TEFxUnN.png)     | ![](https://i.imgur.com/vzhiiJ1.png)   |

### Usage

`docker-compose up`
 
If you want to use the project in its simplest form (no logs being sent to a ZincSearch instance) then just delete the filebeat service from the docker compose. This leaves just the application itself and a redis instance for some light persistence which is required. Fire up the docker compose and the homepage can be viewed at `http://localhost:2495`. 

From the homepage new shelves can be manually crawled. From `http://localhost:2495/available` all books which have been found so far on TheBookshop.ie will be aggregrated. A total cost will be displayed out of €20 because they offer free shipping on orders above that amount. 

For automated checking you first need to set the shelf to crawl which can be done on the bottom of the `http://localhost:2495/available` page. I've got a cron job which gets triggered every 24 hours at 22:00 which looks like this `0 20 * * * curl http://localhost:2945/automatedcheck`. This will then use your specified shelf and crawl it in the backgrund. 

### Configuration

Without any environment variables set the application will be purely contained to its web interface. No alerts will be sent when new books are discovered.

By setting a valid `DISCORD_WEBHOOK_URL` (see [here](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks) on how to setup a webhook for a given channel) booksbooksbooks will notify you through the linked discord channel when a new book from your crawled shelf is available on TheBookshop.ie. A notification can only be sent if a new book is found during a crawl, booksbooksbooks can't know in real time when theBookshop.ie adds a new book but checking once a day works fine.

To run the application with logs being shipped to a [Zincsearch](https://github.com/zinclabs/zinc) (lightweight and simple alternative to elasticsearch/kibana for log aggregration and analysis) instance environment some environment variables are required to be set. I use zincsearch to analyse logs from many of my projects and some extra configuration is required to allow filebeat to authenticate itself. In the base directory create a `.env` file with the following environment variables: `ZINCSEARCH_PASS` and `ZINCSEARCH_INSTANCE_IP` (port included). The docker-compose file will inject these into the filebeat container automatically. Most likely you can ignore these