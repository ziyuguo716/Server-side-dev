"use strict";

const express = require("express");
const morgan = require("morgan");
const mongoose = require("mongoose");
const amqp = require("amqplib/callback_api");
const { channelSchema, messageSchema } = require("./schemas");
const {
    getChannelHandler, postChannelHandler,
    getSpecificChannelHandler, postSpecificChannelHandler, patchSpecificChannelHandler, deleteSpecificChannelHandler,
    channelMemberHandler
} = require("./channel");
const { messageHandler } = require("./message");

const Channel = mongoose.model("Channel", channelSchema);
const Message = mongoose.model("Message", messageSchema);

const mongoEndPoint = "mongodb://mongodb:27017/test"
const rabbitADDR = "amqp://rabbit:5672"
// const mongoEndPoint = "mongodb://localhost:27017/test"
// const rabbitADDR = "amqp://localhost:5672"
const connect = () => {
    mongoose.connect(mongoEndPoint);
}

const app = express();
const port = 4000;
var rabbitChan;

const getRabbitChan = () => {
    return rabbitChan;
}

const RequestWrapper = (handler, SchemeAndDBForwarder) => {
    return (req, res) => {
        handler(req, res, SchemeAndDBForwarder);
    }
}

//add JSON request body parsing middleware
app.use(express.json());
//add the request logging middleware
app.use(morgan("dev"));

//connect to mongoDB
connect();
mongoose.connection.on('error', console.error)
    .on('disconnected', connect)
    .once('open', main);
// mongoose.connection.on('error', console.error)
//     .once('open', main);

//ensure channel 'general' is always in the db
Channel.findOne({ name: "general" }, function (err, result) {
    if (!result) {
        var defaultChannel = new Channel({
            name: "general",
            description: "default public channel for all",
            createdAt: new Date()
        });
        // save model to database
        defaultChannel.save(function (err, channel) {
            if (err) {
                res.status(500).send("Error creating general channel");
                return;
            }
        });
    }
});

app.use("/v1/", function (req, res, next) {
    if (!req.get('X-User')) {
        res.status(401).send("Unauthorized");
        return
    }
    next();
});

app.get("/v1/channels", RequestWrapper(getChannelHandler, { Channel }));
app.post("/v1/channels", RequestWrapper(postChannelHandler, { Channel, getRabbitChan }));

app.get("/v1/channels/:chanid", RequestWrapper(getSpecificChannelHandler, { Channel, Message }));
app.post("/v1/channels/:chanid", RequestWrapper(postSpecificChannelHandler, { Channel, Message, getRabbitChan }));
app.patch("/v1/channels/:chanid", RequestWrapper(patchSpecificChannelHandler, { Channel, getRabbitChan }));
app.delete("/v1/channels/:chanid", RequestWrapper(deleteSpecificChannelHandler, { Channel, Message, getRabbitChan }));

app.post("/v1/channels/:chanid/members", RequestWrapper(channelMemberHandler, { Channel }));
app.delete("/v1/channels/:chanid/members", RequestWrapper(channelMemberHandler, { Channel }));

app.patch("/v1/messages/:msgid", RequestWrapper(messageHandler, { Channel, Message, getRabbitChan }));
app.delete("/v1/messages/:msgid", RequestWrapper(messageHandler, { Channel, Message, getRabbitChan }));

async function main() {
    amqp.connect(rabbitADDR, (error, conn) => {
        if (error) {
            console.log("Error connecting rabbit");
            process.exit(1);
        }
        conn.createChannel((err, chan) => {
            if (err) {
                console.log("Error creating rabbit channel");
                process.exit(1);
            }
            chan.assertQueue("info441", { durable: true });
            rabbitChan = chan;

            // chan.consume("info441", (msg)=> {
            //     console.log(msg.content.toString());
            // },{
            //     noAck:true
            // });

            app.listen(port, "", () => {
                console.log(`server listening ${port}`);
            });
        });
    });


}