# Codecrafter’s http server Project

I have always heard that programmers learn by reading "good" codes. However, I personally always
struggled to do that. Mostly because, understanding why a piece of "good" code is actually good is
not that simple. IMO, you need to write "bad" code to understand why a design decision is "good", why
extracting that little interface was so "good". For that, the best way would be to first try to write
(maybe?) the ugly version yourself. After that, you will actually start to appreciate the beauty of the
"good" code. 

This is exactly the approach I wanted to take while implementing the http server of my own. I know I would
eventually take motivation from the standard library. Yet I did not immediately want to start copying ideas
from the standard library implementation. Rather, I started very simple just to get past a stage, and refactor
it with the best of my abilities, and finally take a look at the standard library for the solution.


## Step 1-5

Simple, just do what they say. At this moment, we can parse a request and get the request path from it. And send back a response.

## Step 6: Respond with Body June 8, 2024

*In this stage, you'll implement the `/echo/{str}` endpoint, which accepts a string and returns it in the response body.*

### Solution

The request format for http is as follows:

```js
GET /index.html HTTP/1.1
Host: localhost:4221
User-Agent: curl/7.64.1
Accept: */*\r\n\r\n
```

It looks like the first line of the request is what we care about in this task. When someone hits our service in the `/echo/abc` path, we will get a request looking something like:

```jsx
GET /echo/abc HTTP/1.1
Host: localhost:4221
User-Agent: curl/7.64.1
Accept: */*\r\n\r\n
```

So, we just need to parse the `path` (which is `/echo/abc`) and return `abc` as the body of the response, along with the `200` status code.

So, we can split the task into two subtasks:

- [x]  Parse the request path and extract the `string` that needs to be echoed.

  Doing this step naively is pretty straightforward. Just match whether the received path follows the regex **`^/echo/\w+$` .** If so, just strip off the `/echo/` prefix, and you will get the desired value.

- [x]  Send the string by setting the `content-type` and `content-length` headers appropriately.

  The `content-type` is simply `text/plain` since we are just sending strings, and calculating the `content-length` is as simple as the length of the returned string. However, be careful about the number of CLRF in the response. Specially make sure there is two CLRFs between the response header section and the response body (one is the trailing CLRF for the last header, another is the section splitter).


Step 6 is complete. The code looks like a mess. Feels like I should refactor it a bit. But, waiting for more concrete reason.

## Read Request Header June 8, 2024

*In this stage, you'll implement the `/user-agent` endpoint, which reads the `User-Agent` request header and returns it in the response body.*

### Solution

At this point, we need more parsing on the coming requests, reading the first so-called `request line` (e.g: GET /index.html HTTP/1.1) is not enough.

So, it feels natural to extract a `Request` struct out, which we will load while parsing the request. The struct `Request` should be:

```go
Request{
	path string,
	headers map[string]string,
	method string // should be enum, but ignoring for the time being
	body []byte
}
```

### Stage: **Return a file June 18, 2024**

Completed the basic returning of the file in response to a file. It may seem difficult at first glance. However, it is really simple once you start implementing it.

There are three parts in total:

1. Read the root file server directory i.e: where the files are to be looked for. Reading the command line argument is good enough for this.
2. Parsing the file name. It is same as the stage `/echo/{str}` stage. So, no biggie.
3. Respond with the actual file content. Till now, we were just working with string response. But this time we will need to send a file of arbitrary binary content. And, for the http client to work appropriately we need to send the `Content-Type` header to `application/octate-stream`.

   > To know more about this MIME type, here is the MDN spec’s relevant part: https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
   >
   >
   > The following two important MIME types are the default types:
   >
   > - `text/plain` is the default value for textual files. A textual file should be human-readable and must not contain binary data.
   > - `application/octet-stream` is the default value for all other cases. An unknown file type should use this type. Browsers are particularly careful when manipulating these files to protect users from software vulnerabilities and possible dangerous behavior.

   However, it is really simple from the implementation angle. Just read the entire file in a byte array and write it as the response body. *Don’t forget to set the `Content-Length` header.* It is simply the length of the byte array.


This concludes adding support to serve a file from the server. So, (maybe 😏) we can use it to server files now? Not so fast! One *possible* future direction is to bench-mark sending a huge file from this server and the one implemented with standard library’s `http` server. One visible shortcoming is loading the entire file into memory instead of loading it in chunk. We should try that in future.

The next stage is to [**Read request body](https://app.codecrafters.io/courses/http-server/stages/qv8).** Before going there, I want to refactor my messed up spaghetti code getting inspired from the standard library.

From where things [stands](https://github.com/ImperishableMe/codecrafters-http-server-go/commit/7276dd33638da4e9b20507938c26b07961a68981) now, we have a few pain points:

1. We cannot write the response body in a streaming manners. Mainly because, we did not expose the `Writer` interface in the [`Response`](https://github.com/ImperishableMe/codecrafters-http-server-go/blob/7276dd33638da4e9b20507938c26b07961a68981/app/http.go#L71-L76) struct. Rather we exposed a `Body` field which expects the entire bytes response at once. In hindsight, the author of a handler needs three things:
   1. A way to write response code
   2. A way to specify some headers
   3. A way to write the body in response to the request in whatever way it wants. However, before the start of the writing body, we must make sure the `Status` and `Headers` are written.
2. As an user of the server, we are blindly passing the `ReadWriteCloser` to the `writeResponse` method every time (e.g: https://github.com/ImperishableMe/codecrafters-http-server-go/blob/7276dd33638da4e9b20507938c26b07961a68981/app/server.go#L52) even though the handler does not use this itself.
3. A lot of code is in the `main` function for no good reason. We should have some sort of `Server` struct to encapsulate the serving request logic.
4. We don’t support HTTP verbs yet.

To fix these issues, let’s have a look at the standard library. We can solve the first two problems by introducing the [`ResponseWriter`](https://pkg.go.dev/net/http#ResponseWriter) from standard library.



Expected Supported Spec:

```go
server := &Server{}
server.Register("GET /files/{filename}", func(r *Request, w ResponseWriter))
```

`Register` method on `Server` struct will take a pattern and a handler function.

pattern: The pattern should be `<http verb> /literal/{pathParam}` format.

Supported verbs: `Get` and `Post`

For the matched pattern, we won’t support most specific pattern matching logic. The resolution rule will be simple:

The handlers will be checked one by one in the order of their registration. The first one that matches will handle the request, and there won’t be a way to delegate the handling to a later handler. If no match is found, just handle with `404 Not Found`.

## Making the Request Body Streaming!

faced a really critical bug! Using the partially (after scanning the request line, and headers) read `net.Conn` as the request.Body. The root of the bug was the `scanner` uses a buffer and reads up everything (*48k bytes*). So, when the handler tries to read the body, there is nothing left on the stream.