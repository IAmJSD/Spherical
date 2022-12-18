# Spherical

[My Kofi](https://ko-fi.com/jakemakesstuff)

Spherical is a work in progress decentralized chat application. Think like Discord but with the decentralisation of Mastodon.

## Current Progress
This list will grow as I think of things that we need to release:

- [X] Hash Validators
  - [X] Local caching
  - [X] Local validation
  - [X] Accessing /spherical.pub to check key
  - [X] Check trusted nodes to check expected results
  - [X] Skip nodes we have already tried recursively to prevent DDoS attack
- [ ] Integration Tests
  - [X] Structure
  - [ ] Implementation
- [ ] Unit Tests
  - [ ] Setup Helpers
  - [ ] Main Tests
  - [X] Helper Tests
- [ ] GitHub Actions
- [X] Database Structure
- [X] Migrations
- [X] Event Dispatching
- [X] Event Handling
- [X] E-mail Support
- [ ] E-mail Verification
- [X] Access Token Management
- [X] Internationalisation Engine
- [X] Frontend Structure
- [X] Shared Frontend Logic
  - [X] Internationalisation Loader
  - [X] Localisation for alt text and standard content
  - [X] Streaming in markdown loader
  - [X] Basic helpers (Notifications, buttons, loading, flexbox)
- [ ] User Registration
  - [X] Backend functions
  - [ ] API routes
  - [ ] Frontend
- [X] Out of box experience
  - [X] Setup key fetching
  - [X] S3 configuration (amd tests)
  - [X] E-mail configuration (and tests)
  - [X] Server Information
  - [X] Owner User Setup
  - [ ] Hash Validators
  - [X] Install completion
- [X] Router switching between OOBE and main app
- [X] User Login
  - [X] Backend functions
  - [X] API routes
  - [ ] Frontend
- [ ] Gateway
  - [X] Cross Node Login
  - [X] Local Login
  - [X] Local Authentication
  - [X] Structure
  - [X] Heartbeats
  - [X] Ready Payload
  - [ ] Core Functionality
  - [ ] Documentation
- [X] Auth API abstraction
- [ ] API V1 routes (TBD)
- [ ] Application frontend (TBD)

## Developing the application
To develop for Spherical, you will want to have both node.js and Go installed on your computer. You will also want to have a copy of Postgres and Redis locally to make it easier than trying to deal with some external database servers during development.

You will also want access to a e-mail server (can be the same as production, doesn't matter too much), and a S3 bucket (this should be separate from your production instance). From here, you will want to run `yarn` to setup the JS packages. From here, you will want to run `yarn run start:dev` to start the development server. When in development mode, Spherical will use the hard disk to grab elements, but when built in production mode, it will use the embedded files for performance and simplicity reasons.

You can just run through the setup in dev mode like you would on a production install, but it is worth noting unless you use something like Cloudflare Argo Tunnels, you will not be able to do anything cross node.

If you wish to try building for all platforms, you can use `yarn build`. Be warned this will take a while and use a lot of CPU, but after that you will have builds for a lot of platforms. Spherical basically targets all architectures of Windows, Linux, and FreeBSD supported by Go. If you want the full list of architectures Spherical is built for, consult the list inside `scripts/buildBinaries.js`.

## Distributed Binaries
This application has some binary files distributed with it for use with the frontend. These are the following:
- **Roboto**: Licensed under [Apache-2.0](https://fonts.google.com/specimen/Roboto/about). No changes made.
- **Twemoji:** Licensed under [CC-BY 4.0](https://creativecommons.org/licenses/by/4.0/). No changes made.

Note these binaries fall under their respective licenses.
