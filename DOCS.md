## Documentation


1. [Introduction](#introduction)
2. [Authentication](#authentication)
3. [Endpoints](#endpoints)
    1. [Account Creation](#account)
    2. [Balance Check](#balance)
    3. [Verify](#verify)
    4. [Get Logs](#logs)
    5. [Transaction](#transaction)


### Introduction
"Schule als Staat" is a German school initiative where students emulate the operation of a state, complete with its own economy, legal system, and cultural events. This project is designed to equip students with hands-on experience of political, economic, and social processes. 

In an increasingly digital world (where tax fraud is a concern 🥲), we've decided to adapt and provide a comprehensive digital currency solution. Our software is open-source and freely available on GitHub. 

To facilitate interactions with the banking system, we offer a REST API that can be accessed via our mobile app or web interface. This documentation provides detailed information about the API, enabling the software to be utilized for other "Schule als Staat" projects.

### Authentication
The API is guarded by a Bearer Token system, this measure to ensure a simple access authorization, it is highly recommended to change the token in code to a secret password ensuring that only authorized clients may access a specific instance. The token needs to be provided in the header through the following:
``` Authorization : Bearer YOUR TOKEN ```

### Endpoints

#### Account Creation (POST) <a name="account"></a>
This endpoint is used to create a new account for a user. The user must provide a unique username and password. The account will be created with a balance of 0.0. 

``` /addAccount```

Request Body:

```json
{
    "name": "gunter", //provide a name, it will only be used for verification, the account will get a unique id
    "pin": "1234" //provide a pin, it will be used to verify the account
}
```

#### Balance Check (POST) <a name="balance"></a>
This endpoint is used to check the balance of an account. The user must provide their account ID and pin. 

``` /balanceCheck```

Request Body:

```json
{
    "acc1": "XrfG4m...", //provide the account id
    "pin": "1234" //provide the pin
}
```

#### Verify (POST) <a name="verify"></a>
This endpoint is used to verify an account. The user must provide their account ID and pin. The Route checks if the account exists and if the pin is correct.

``` /verify```

Request Body:

```json
{
    "name": "XrfG4m...", //provide the account id
    "pin": "1234" //provide the pin
}
```

#### Get Logs (POST) <a name="logs"></a>
This endpoint is used to get the logs of an account. The user must provide their account ID and pin. The Route returns all transactions where the account is involved.

``` /getLogs```

Request Body:

```json
{
    "name": "XrfG4m...", //provide the account id
    "pin": "1234" //provide the pin
}
```

#### Transaction (POST) <a name="transaction"></a>
This endpoint is used to make a transaction between two accounts. The user must provide their account ID, the account ID of the recipient, the amount, and the pin. The Route checks if the account exists and if the pin is correct. If the account has enough balance, the transaction will be executed.

``` /pay```

Request Body:

```json
{
    "acc1": "XrfG4m...", //provide the account id
    "acc2": "asGkM3...", //provide the account id of the recipient
    "amount": 10.0, //provide the amount
    "pin": "1234" //provide the pin
}
```