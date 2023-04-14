# Day Trading Application

UVic course project for SENG 468.

This project consisted of making a scalable stock trading app where users can create accounts, add virtual funds and complete trading transactions.

## Getting Started

These instructions will cover usage information

### Prerequisities

In order to run this project you'll need docker installed.

- [Windows](https://docs.docker.com/desktop/install/windows-install/)
- [OS X](https://docs.docker.com/desktop/install/mac-install/)
- [Linux](https://docs.docker.com/desktop/install/linux-install/)

## To run

Clone the repo

    git clone https://github.com/andreaRA99/SENG468.git

Ensure that the Docker Daemon is running.

In the project root directory run the following command

    docker compose up --build

Visit

    `web-server`

To stop the containers

    docker-compose down

# Supported Functionality

## Command-Line Interface

### To run

Start a new terminal instance in the project root directory, separate from the instance running the dockerized appplication, and run the following command

    go run cli.go

This will give a summary of the commands available to execute.

To view a commad's options run

    go run cli.go [command] -h

or

    go run cli.go [command] --help

#### Example - Getting command help

    go run cli.go r -h

#### Example - Running CLI command

    go run cli.go r --fl ./resources/user1.txt

#### Example - Running user command via CLI

    go run cli.go e --cmd set_buy_amount --id test_user --stock aaa --amount 500

## User Commands

The application supports the execution of the following commands

- ADD \[userid, amount\]
- QUOTE \[userid, StockSymbol\]
- BUY \[userid, StockSymbol, amount\]
- COMMIT_BUY \[userid\]
- CANCEL_BUY \[userid\]
- SELL \[userid, StockSymbol, amount\]
- COMMIT_SELL \[userid\]
- CANCEL_SELL \[userid\]
- SET_BUY_AMOUNT \[userid, StockSymbol, amount\]
- CANCEL_SET_BUY \[userid, StockSymbol\]
- SET_BUY_TRIGGER \[userid, StockSymbol, amount\]
- SET_SELL_AMOUNT \[userid, StockSymbol, amount\]
- SET_SELL_TRIGGER \[userid, StockSymbol, amount\]
- CANCEL_SET_SELL \[userid, StockSymbol\]
- DUMPLOG \[userid, filename\]
- DUMPLOG \[filename\]
- DISPLAY_SUMMARY \[userid\]

# Authors

**Andrea Ramirez**

**Dennis Arimi**

**Mateo Moody**

**Radu Ionescu**

**Ryan O'Hara**
