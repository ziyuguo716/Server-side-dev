const Schema = require("mongoose").Schema;

const channelSchema = new Schema({
    name: { type: String, required: true, unique: true },
    description: String,
    private: Boolean,
    members: [],
    createdAt: { type: Date, required: true },
    creator: Object,
    editedAt: Date
})

const messageSchema = new Schema({
    channelID: { type: String, required: true },
    body: String,
    createdAt: { type: Date, required: true },
    creator: Object,
    editedAt: Date
})


module.exports = { channelSchema, messageSchema }