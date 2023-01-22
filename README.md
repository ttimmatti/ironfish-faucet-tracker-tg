# ironfish-faucet-tracker-tg

Checks if IronFish Faucet is working or not and updates the message in the chat if state changes.

To use:
1. change _bot_token value to your bot's api token
2. change _chat_id value to the id of the chat you want the bot to send this message to

The script will send a new message in the chat, first message will always say that faucet is not working
Then every 3 minutes it will try to update the message. If state of the faucet has changed the message will be updated, if not same text will remain.
