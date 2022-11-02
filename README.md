# booksbooksbooks

I use GoodReads for my book reccomendations and I buy my books mostly from TheBookshop.ie. I made this project to be able to automatically check if books on my "shopping list" are available on TheBookshop. Still in active development but hoping to make a 1.0 release before the end of November 2022.

|      |  |
| ----------- | ----------- |
| ![](https://i.imgur.com/TEFxUnN.png)     | ![](https://i.imgur.com/vzhiiJ1.png)   |

### Usage

`docker-compose up`
 
If you want to use the project in its simplest form (no logs being sent to a zincsearch instance) then just delete the filebeat service from the docker compose. This leaves just the application itself and a redis instance for some light persistence. Fire up the docker compose and the homepage can be viewed at `http://localhost:2495`. 

From the homepage new shelves can be manually crawled. From `http://localhost:2495/available` all books which have been found so far on TheBookshop will be aggregrated. A total cost will be displayed out of €20 because they offer free shipping on orders above that amount. 

For automated checking of my "shopping list" shelf the input on this page is set to that GoodReads shelf. I've got a cron job which triggers every 24 hours at 22:00 which looks like this `0 20 * * * curl http://localhost:2945/automatedcheck` and executed the `/automatedcheck` endpoint. Notifications will most likely be supported through discord webhooks as I find those to be the handiest.

### Configuration

To run the application without log shipping to a [Zincsearch](https://github.com/zinclabs/zinc) (lightweight and simple alternative to elasticsearch/kibana for log aggregration and analysis) instance no environment variables are required to be set. I use zincsearch to analyse logs from many of my projects and some extra configuration is required to allow filebeat to authenticate itself. In the base directory create a `.env` file with the following environment variables: `ZINCSEARCH_PASS` and `ZINCSEARCH_INSTANCE_IP` (port included). The docker-compose file will inject these into the filebeat container automatically