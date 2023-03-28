# Prometheus exporter for Skoda vehicles

This Prometheus exporter connects to Skoda cloud services to receive vehicle metrics.

![build](https://github.com/chr4/enyaq_exporter/workflows/build/badge.svg)

It's currently retrieving the following values:

- Eletric vehicle range
- Eletric vehicle state of charge (as integers)
- Vehicle status
- Charging finish time
- Odometer

It should work with all Skoda EVs, but is mainly tested with the Skoda Enyaq.

All the heavy lifting is done by [evcc libraries](https://github.com/evcc-io/evcc/).
