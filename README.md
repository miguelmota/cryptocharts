# cryptodash

> Cryptocurrency information and charts displayed in a dashboard from your terminal.

<img src="./screenshot.png" width="750">

# Install

Make sure to have [golang](https://golang.org/) installed, then do:

```bash
go get -u github.com/miguelmota/cryptodash/cryptodash
```

# Usage

```bash
$ cryptodash {cryptocurrency} {chart_date_range ie. 1h | 1d | 2d | 7d | 30d | 2w | 1m | 3m | 1y} [primary_color ie. green | cyan | magenta | red | yellow | white]
```

Here's an example of getting latest [Ethereum](https://www.ethereum.org/) info, and chart data for the last 30 days.

```bash
$ cryptodash ethereum 30d
```

(output is screenshot above)

## FAQ

- Q: Where is the data from?

  - A: The data is from [Coin Market Cap](https://coinmarketcap.com/).

- Q: What coins does this support?

  - A: This supports any coin listed on [Coin Market Cap](https://coinmarketcap.com/).

- Q: How often is the data polled?

  - A: Data gets polled once every minute.


# License

MIT
