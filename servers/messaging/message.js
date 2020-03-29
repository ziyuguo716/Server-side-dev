//messageHandler edit and delete message by creator
const messageHandler = async (req, res, { Channel, Message, getRabbitChan }) => {
    const msgid = req.params.msgid;
    if (!msgid) {
        res.status(400).send("Message id not provided");
        return;
    }
    const msg = await Message.findOne({ _id: msgid });
    if (!msg) {
        res.status(400).send("Message not found");
        return;
    }

    var memberIDs = [];
    const channel = await Channel.findOne({ _id: msg.channelID });
    if (!channel) {
        res.status(404).send("No channel linked to this message");
        return;
    }
    if (channel.private) {
        for (i = 0; i < channel.members.length; i++) {
            memberIDs.push(channel.members[i].id)
        }
    }

    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());
    if (!isCreatorOf(msg, user)) {
        res.status(403).send("Not a creator of this message");
        return;
    }
    const { text } = req.body;
    if (req.method == 'PATCH') {
        Message.findByIdAndUpdate(msgid,
            {
                "body": text,
                "editedAt": new Date()
            },
            { new: true, safe: true, upsert: true },
            function (err, data) {
                if (err) {
                    res.status(500).send("There was an error updating message");
                    return;
                }
                //send new channel JSON to rabbitMQ
                let rabbitChan = getRabbitChan();
                rabbitChan.sendToQueue("info441", Buffer.from(JSON.stringify(
                    {
                        type: "message-update",
                        message: data,
                        userIDs: memberIDs
                    }
                )));
                res.setHeader("Content-Type", "application/json");
                res.status(201).json(data);
            });
    }
    if (req.method == 'DELETE') {
        await msg.remove()
        //send new channel JSON to rabbitMQ
        let rabbitChan = getRabbitChan();
        rabbitChan.sendToQueue("info441", Buffer.from(JSON.stringify(
            {
                type: "message-delete",
                messageID: msgid,
                userIDs: memberIDs
            }
        )));
        res.status(200).send("Success removing message");
    }

}

function isCreatorOf(msg, user) {
    if (msg.creator.id !== user.id) {
        return false;
    }
    return true;
}


module.exports = { messageHandler };