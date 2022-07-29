# rate-interval-rule
Checks that every target with a `rate`, `irate` or `increase` function uses `$__rate_interval` for the range of the data to process.

# Best Practice
In short, this ensures that there is always a sufficient number of data points to calculate a useful result. A detailed description can be found in [this Grafana blog post](https://grafana.com/blog/2020/09/28/new-in-grafana-7.2-__rate_interval-for-prometheus-rate-queries-that-just-work/)

# Possible exeptions
There may be cases where one deliberately wants to show the rate or increase over a fixed period of time, such as the last 24hr etc. In those cases you may wish to create a lint exclusion for this rule.