//handle GET/POST '/v1/channels'
const getChannelHandler = async (req, res, { Channel }) => {
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());
    try {
        res.set("Content-Type", "application/json");
        const channels = await Channel.find();
        var channelResults = [];
        for (var channel of channels) {
            if (!channel.private) {
                channelResults.push(channel);
            } else {
                for (var member of channel.members) {
                    if (user.id == member.id) {
                        channelResults.push(channel);
                        continue;
                    }
                }
            }
        }
        res.json(channelResults);
    } catch (e) {
        res.status(500).send("There was an error getting channels");
    }
}

const postChannelHandler = async (req, res, { Channel, getRabbitChan }) => {
    console.log("posting new channel......")
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());

    var { cname, cdescription, cprivate, cmembers } = req.body;
    var memberIDs = [];
    if (!cname) {
        res.status(400).send("Must provide a channel name");
        return;
    }
    if (!cprivate) {
        cprivate = false;
        cmembers.length = 0;
    } else {
        for (i = 0; i < cmembers.length; i++) {
            memberIDs.push(cmembers[i].id)
        }
    }
    let errFound = false;
    await Channel.findOne({ "name": cname }, function (err, result) {
        if (err) {
            res.status(500).send("There was an error getting channel");
            return;
        }
        if (result) {
            errFound = true;
        }
    });
    if (errFound) {
        res.status(400).send("Must provide a unique channel name");
        return;
    }
    const newChannel = {
        name: cname,
        description: cdescription,
        private: cprivate,
        members: cmembers,
        createdAt: new Date(),
        creator: user
    };

    const query = new Channel(newChannel);
    query.save((err, newChan) => {
        if (err) {
            res.status(500).send("Error saving new channel");
            return;
        }

        //send new channel JSON to rabbitMQ
        let rabbitChan = getRabbitChan();
        rabbitChan.sendToQueue("info441", Buffer.from(JSON.stringify(
            {
                type: "channel-new",
                channel: newChan,
                userIDs: memberIDs
            }
        )));

        res.setHeader("Content-Type", "application/json");
        res.status(201).json(newChan);
    });
}


const getSpecificChannelHandler = async (req, res, { Channel, Message }) => {
    console.log("getting channel.....")
    const chanid = req.params.chanid;
    const channel = await Channel.findOne({ _id: chanid });
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());
    //channel requested not found
    if (!channel) {
        res.status(404).send("Invalid channel");
        return;
    }

    if (channel.private) {
        if (!isMemberOf(channel, user)) {
            res.status(403).send("Not a member of this private channel");
            return;
        }
    }

    const msgid = req.query.before;
    res.setHeader("Content-Type", "application/json");
    if (msgid) {
        const msgBefore = await Message.find({ channelID: channel._id }).
            sort({ createdAt: -1 }).
            where("_id").lt(msgid)
            .limit(100);
        res.status(201).json(msgBefore);
        return;
    }
    const msg = await Message.find({ channelID: channel._id }).sort({ createdAt: -1 }).limit(100);
    res.status(201).json(msg);
}

const postSpecificChannelHandler = async (req, res, { Channel, Message, getRabbitChan }) => {
    console.log("posting channel.....")
    const chanid = req.params.chanid;
    const channel = await Channel.findOne({ _id: chanid });
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());
    var memberIDs = [];
    //channel requested not found
    if (!channel) {
        res.status(404).send("Invalid channel");
        return;
    }
    if (channel.private) {
        if (!isMemberOf(channel, user)) {
            res.status(403).send("Not a member of this private channel");
            return;
        }

        for (i = 0; i < channel.members.length; i++) {
            memberIDs.push(channel.members[i].id)
        }
    }
    const { msg } = req.body;
    const newMsg = {
        channelID: chanid,
        body: msg,
        createdAt: new Date(),
        creator: user
    }
    const query = new Message(newMsg);
    query.save((err, newMessage) => {
        if (err) {
            res.status(500).send("Error saving new message");
            return;
        }
        //send new channel JSON to rabbitMQ
        let rabbitChan = getRabbitChan();
        rabbitChan.sendToQueue("info441", Buffer.from(JSON.stringify(
            {
                type: "message-new",
                message: newMessage,
                userIDs: memberIDs
            }
        )));
        res.setHeader("Content-Type", "application/json");
        res.status(201).json(newMessage);
    });
}

const patchSpecificChannelHandler = async (req, res, { Channel, getRabbitChan }) => {
    console.log("Patching channel.....")
    const chanid = req.params.chanid;
    const channel = await Channel.findOne({ _id: chanid });
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());
    //channel requested not found
    if (!channel) {
        res.status(404).send("Invalid channel");
        return;
    }

    if (!isCreatorOf(channel, user)) {
        res.status(403).send("Not a creator of this channel");
        return;
    }
    const { newname, newdescription } = req.body;
    var query = { _id: chanid };
    var update = { name: newname, description: newdescription, editedAt: new Date() };
    if (!newdescription) {
        update = { name: newname, editedAt: new Date() };
    }
    var options = { new: true };
    await Channel.findOneAndUpdate(query, update, options, function (err, newChan) {
        if (err) {
            res.status(500).send("Error updating channel");
            return;
        }
        var memberIDs = [];
        if (newChan.private) {
            for (i = 0; i < newChan.members.length; i++) {
                memberIDs.push(newChan.members[i].id)
            }
        }
        //send new channel JSON to rabbitMQ
        let rabbitChan = getRabbitChan();
        rabbitChan.sendToQueue("info441", Buffer.from(JSON.stringify(
            {
                type: "channel-update",
                channel: newChan,
                userIDs: memberIDs
            }
        )));
        res.setHeader("Content-Type", "application/json");
        res.status(201).json(newChan);
    });
}

const deleteSpecificChannelHandler = async (req, res, { Channel, Message, getRabbitChan }) => {
    const chanid = req.params.chanid;
    const channel = await Channel.findOne({ _id: chanid });
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());
    //channel requested not found
    if (!channel) {
        res.status(404).send("Invalid channel");
        return;
    }

    if (!isCreatorOf(channel, user)) {
        res.status(403).send("Not a creator of this channel");
        return;
    }
    await Message.deleteMany({ channelID: chanid }, function (err) {
        if (err) {
            res.status(500).send("Error deleting message");
            return;
        }
    });
    var memberIDs = [];
    if (channel.private) {
        for (i = 0; i < channel.members.length; i++) {
            memberIDs.push(channel.members[i].id)
        }
    }
    await channel.remove();
    //send new channel JSON to rabbitMQ
    let rabbitChan = getRabbitChan();
    rabbitChan.sendToQueue("info441", Buffer.from(JSON.stringify(
        {
            type: "channel-delete",
            channelID: chanid,
            userIDs: memberIDs
        }
    )));
    res.status(200).send("Success delete this channel");
}


const channelMemberHandler = async (req, res, { Channel }) => {
    const chanid = req.params.chanid;
    const channel = await Channel.findOne({ _id: chanid });
    var buff = Buffer.from(req.get('X-User'), 'base64');
    const user = JSON.parse(buff.toString());

    const { id } = req.body;
    if (!id) {
        res.status(400).send("Must provide a user id");
        return;
    }
    //channel requested not found
    if (!channel) {
        res.status(404).send("Invalid channel");
        return;
    }
    if (!isCreatorOf(channel, user)) {
        res.status(403).send("Not a creator of this channel");
        return;
    }
    if (req.method == 'DELETE') {
        await Channel.findByIdAndUpdate(chanid,
            {
                $pull: { "members": { "id": id } }
            },
            { new: true, safe: true, upsert: true },
            function (err) {
                if (err) {
                    res.status(500).send("There was an error remove member to channel");
                    return;
                }
            });
        res.status(200).send("Success removing user to the channel");
    }
    if (req.method == 'POST') {
        Channel.findByIdAndUpdate(chanid,
            {
                $push: { "members": req.body }
            },
            { new: true, safe: true, upsert: true },
            function (err) {
                if (err) {
                    res.status(500).send("There was an error add member to channel");
                    return;
                }
            });
        res.status(201).send("Success adding user to the channel");
    }
}

//check user is a member of private channel
function isMemberOf(channel, user) {
    if (channel.private) {
        const members = channel.members;
        for (member of members) {
            if (member.id == user.id) {
                return true;
            }
        }
    }
    return false;
}

function isCreatorOf(channel, user) {
    if (channel.creator.id !== user.id) {
        return false;
    }
    return true;
}

module.exports = {
    getChannelHandler, postChannelHandler,
    getSpecificChannelHandler, postSpecificChannelHandler, patchSpecificChannelHandler, deleteSpecificChannelHandler,
    channelMemberHandler
};