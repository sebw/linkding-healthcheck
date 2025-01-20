# LDHC: Linkding Health Check and Duplicate Finder

[Linkding](https://github.com/sissbruecker/linkding/) is an amazing bookmark manager!

At the time of writing this README (Jan 14th, 2025), my Linkding instances has 5359 bookmarks.

My first ever recorded bookmark was on Jan 13th.... 2006. At 12:27. That's 19 years and a day!

I'm showing my age here but I digress. 

My bookmarks have survived the ages and tools (Delicious, Pinboard, Shaarli until Linkding) but not necesarily the pages themselves.

The internet comes and goes and it can be a good thing to know if a page went 404, and somehow mark it as dead in Linkding.

This request https://github.com/sissbruecker/linkding/issues/68 from 2021 is exactly asking for that but I understand the challenges in implementing something like that. I also like the lightweight approach of Linkding and appreciate it doesn't become a bloated piece of software.

This is why I decided to implement a quick and dirty URL health checker outside of Linkding and take advantage of the tag systems.

This app also checks if your Linkding contains exact duplicates, this can happen at import time because Linkding doesn't check duplicates. This topic is discussed [here](https://github.com/sissbruecker/linkding/issues/751).

You just need to run the following Docker image on a regular basis and a tag `@HEALTH_HTTP_<code>` will be assigned in case the link is no longer valid.

In case of DNS issue `@HEALTH_DNS` is assigned and for other cases (TLS, etc.) `@HEALTH_other`.

Tags are prefixed with `@` so they are at the top of the list.

In between two runs:

- if a site comes back to life, the old health tag(s) is/are removed
- if the error changes (from 403 to 404 for example), both tags are kept, so you see the different errors the site has been dealing with since it went dead

The code is mimicking a Chrome browser to limit false positive, but many 403 are not actual errors.

>[!WARNING]
>Backup your Linkding instance before running ldhc.

You just need to provide your Linkding API endpoint and token as parameters to the following Docker command:

```bash
docker run -it --rm --name=ldhc -e API_TOKEN="ABCDEF" -e API_URL="https://your.linkding.example.com/api/bookmarks" ghcr.io/sebw/ldhc:latest
```

Output:

```
Fetched 100 bookmarks so far...
Fetched 200 bookmarks so far...
[...]
Fetched 5300 bookmarks so far...
Fetched 5359 bookmarks so far...
Total bookmarks fetched: 5359
[...]
Bookmark 5417 ✅ https://www.example.com/blog/2024-02-18
Bookmark 5395 ❌ HTTP_403 https://www.example.com/de-en/listing/123/blah
[...]
Duplicate Links:
URL: https://blah.example.com/ - Duplicate in bookmarks: [6 5]
```
