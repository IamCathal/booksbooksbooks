# booksbooksbooks

I use GoodReads for my book reccomendations and I buy my books mostly from TheBookshop.ie. I made this project to be able to automatically check if books on my "shopping list" are available on TheBookshop. Still in active development but hoping to make a 1.0 release before the end of November 2022.

|      |  |
| ----------- | ----------- |
| ![](https://i.imgur.com/TEFxUnN.png)     | ![](https://i.imgur.com/vzhiiJ1.png)   |

### Usage

* `docker-compose up`
* Go to [http://localhost:2945/settintgs](http://localhost:2945/settings) to set the automated crawl shelf URL, time and discord webhook URL (if you want alerts)
* Start a manual crawl on the homepage or wait for an automated crawl to start on its own (if you've set a shelf URL and time for it to use)

From the homepage new shelves can be manually crawled. From [http://localhost:2945/available](http://localhost:2945/available) all books which have been found so far on TheBookshop.ie will be aggregrated. A total cost will be displayed out of €20 because they offer free shipping on orders above that amount. 

For automated checking you first need to set the (goodreahds) shelf to crawl which can be done onthe [http://localhost:2945/settingsf](http://localhost:2945/settings) page. This will then use your specified shelf and crawl it in the backgrund when the time comes

### Configuration

To run the application with logs being shipped to a [Zincsearch](https://github.com/zinclabs/zinc) (lightweight and simple alternative to elasticsearch/kibana for log aggregration and analysis) instance environment some environment variables are required to be set. I use zincsearch to analyse logs from many of my projects and some extra configuration is required to allow filebeat to authenticate itself. In the base directory create a `.env` file with the following environment variables: `ZINCSEARCH_PASS` and `ZINCSEARCH_INSTANCE_IP` (port included). The docker-compose file will inject these into the filebeat container automatically. Most likely you can ignore these