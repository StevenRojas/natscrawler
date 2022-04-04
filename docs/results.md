## Storage

The URL information and statistics are stored at Postgres as follows:

![Alt text](storage.png?raw=true "URL info storage")

The `stats` field contains this information:
```json
{
  "collector_id": 5,
  "waiting": {
    "start_at": "2022-04-03T18:33:23.9539096Z",
    "end_at": "2022-04-03T18:33:24.1333931Z",
    "duration": 179
  },
  "collector": {
    "start_at": "2022-04-03T18:33:24.1334302Z",
    "end_at": "2022-04-03T18:33:28.9227761Z",
    "duration": 4789
  }
}
```

## API execution
This approach uses the API URL available to avoid run a chrome instance and parse the HTML content. Instead, it is use a REST call.
This is much faster approach.

![Alt text](api.png?raw=true "API execution") 

With 10 `crawler instances` the time needed to parse 22767 URLs was around 16 minutes
```text
Started at: 2022-04-03T18:02:27.1446368Z
Finished at: 2022-04-03T18:17:33.2472575Z
```

With 30 `crawler instances` the time needed to parse 22767 URLs was around 4 minutes
```text
Started at: 2022-04-03T18:30:14.9505611Z
Finished at: 2022-04-03T18:36:16.8661423Z
```

## Chromedp execution
This approach uses `zenika/alpine-chrome:latest` image which runs a headless mode Chrome instance used by `chromedp` to do the crawler.

![Alt text](chromedp.png?raw=true "Chromedp execution")

As displayed in the image above, there are 40 instances of `chrome` image and 10 `crawler` instances. With this configuration, 
the time needed to parse 600 URLs was around 20 minutes
```text
Started at: 2022-04-03T13:21:59.550617274Z
Finished at: 2022-04-03T13:44:01.452587753Z
```

## Console

![Alt text](console.png?raw=true "Console")

1. The `reader` console where some simulated errors are displayed, when the URL is invalid or when the `listener service` reject the request.
2. The `listener` console where a counter of received URLs is displayed
3. The `crawler` console that displays the logs of incoming requests and processed URLs