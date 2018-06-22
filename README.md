GraphQL Multipart Middleware
============================

[![](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://godoc.org/github.com/lucassabreu/graphql-multipart-middleware)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1f0199e1dd364abcae45fd1f3de3cc25)](https://www.codacy.com/app/lucassabreu/graphql-multipart-middleware?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=lucassabreu/graphql-multipart-middleware&amp;utm_campaign=Badge_Grade)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flucassabreu%2Fgraphql-multipart-middleware.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Flucassabreu%2Fgraphql-multipart-middleware?ref=badge_shield)
[![Build Status](https://travis-ci.org/lucassabreu/graphql-multipart-middleware.svg?branch=master)](https://travis-ci.org/lucassabreu/graphql-multipart-middleware)
[![Coverage Status](https://coveralls.io/repos/github/lucassabreu/graphql-multipart-middleware/badge.svg?branch=master)](https://coveralls.io/github/lucassabreu/graphql-multipart-middleware?branch=master)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flucassabreu%2Fgraphql-multipart-middleware.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Flucassabreu%2Fgraphql-multipart-middleware?ref=badge_shield)

This packages provide a implementation of the graphql multipart request spec created by [@jaydenseric](https://github.com/jaydenseric) to provide support for handling file uploads in a GraphQL server, [click here to see the spec](https://github.com/jaydenseric/graphql-multipart-request-spec).

Using the methods `graphqlmultipart.NewHandler` or `graphqlmultipart.NewMiddlewareWrapper` you will be abble to wrap your GraphQL handler and so every request made with the `Content-Type`: `multipart/form-data` will be handled by this package (using a provided GraphQL schema), and other `Content-Types` will be directed to your handler.

The package also provide a scalar for the uploaded content called `graphqlmultipart.Upload`, when used it will populate your `InputObjects` or arguments with a `*multipart.FileHeader` for the uploaded file that can be used inside your queries/mutations.



## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flucassabreu%2Fgraphql-multipart-middleware.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Flucassabreu%2Fgraphql-multipart-middleware?ref=badge_large)
