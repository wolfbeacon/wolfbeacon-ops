const AWS = require('aws-sdk');
const SlackBot = require('slackbots');
const config = require('./config');
const colors = require('colors');

const routines = {
    'NewPrivateBucketRoutine': require('./routines/NewPrivateBucketRoutine'),
    'ListBucketsRoutine': require('./routines/ListBucketsRoutine')
};

let authorizedUsers = {};

const User = require('./app/User');

let bot = new SlackBot({
    token: config.botKey,
    name: 'Operations Bot'
});

let defaultParams = {
    icon_emoji: ':computer:'
};

let botId = null;
let channelId = null;

console.log("Welcome to WolfBeacon internal operations bot.".bold.green);
console.log("Author: Jun Zheng (me at jackzh dot com, Site Reliability Engineer)".bold);
console.log("----------------------------------------");
console.log("Initializing operations bot _(:3 」∠ )_");


bot.on('start', function() {
    console.log("RTM connection to Slack established ヽ(●´∀`●)ﾉ");
    bot.postMessageToChannel(config.botChannel, 'Operations bot online!', defaultParams);
    bot.getUsers().then((data) => {
        for(let index in data.members){
            let user = data.members[index];
            if(User.isAuthorized(user.profile.email)){
                console.log(("Authorized user found " + user.profile.email + " " + user.name + " " + user.id).green.bold);
                authorizedUsers[user.id] = {
                    name: user.name,
                    authorizations: User.find(user.profile.email).authorizations,
                    email: user.profile.email,
                    id: user.id
                };
            } else {
                console.log(("Unauthorized user found " + user.profile.email + " " + user.name + " " + user.id).red.bold);
            }
            if(user.name === config.botName){
                botId = user.id;
                console.log(("Bot ID found " + botId).green.bold);
            }
        }
    });
    bot.getChannel(config.botChannel).then((data) => {
        channelId = data.id;
        console.log(("Channel ID found " + channelId).green.bold);
    });
});

bot.on('message', function(data){
    if(data.type == 'message' && data.channel == channelId && data.user != botId && !data.bot_id){
        console.log(new Date() + " Received new message from " + data.user + ", triggering response.");

        if(/^!op .*$/.test(data.text)){
            let message = data.text.replace('!op ', '');
            if(/^routines$/.test(message)){
                // List routines available
                let response = "List of routines available (type `!op my routines` to see a list of routines you can execute):\n";
                for(let index in routines){
                    response += index + '\n';
                }
                bot.postMessageToChannel(config.botChannel, response + ' <@' + data.user + '>', defaultParams);

            // List executable routines
            } else if(/^my routines$/.test(message)){
                let response = "List of routines you can execute:\n";
                for(let index in authorizedUsers[data.user].authorizations){
                    response += authorizedUsers[data.user].authorizations[index] + '\n';
                }
                bot.postMessageToChannel(config.botChannel, response + ' <@' + data.user + '>', defaultParams);
            } else if(/^execute .*$/.test(message)){
                let routineName = message.split(" ")[1];
                console.log("Requested routine execute of " + routineName);
                let args = message.split(" ");
                args.shift(); args.shift();
                if(User.canDo(authorizedUsers, data.user, routineName)){
                    console.log("Authorized");

                    bot.postMessageToChannel(config.botChannel, 'Your request has been received, now executing routine. <@' + data.user + '>', defaultParams);

                    let routineConfig = routines[routineName].parseArgs(args);
                    let routineInstance = new routines[routineName](routineConfig);

                    routineInstance.process((result) => {
                        let finalResponse = "Routine execution successful.\n";
                        finalResponse += "```" + JSON.stringify(result, null, 2) + "```";
                        finalResponse += '<@' + data.user + '>';
                        bot.postMessageToChannel(config.botChannel, finalResponse, defaultParams);
                    }, (err) => {

                    });
                } else {
                    bot.postMessageToChannel(config.botChannel, 'You are not authorized to execute that routine, or it does not exist. <@' + data.user + '>', defaultParams);
                    console.log("Not authorized");
                }
            }

        } else {
            console.log("Not a command");
        }
    }
});