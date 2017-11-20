const config = require('../config');

class User {
    constructor(){

    }

    static isAuthorized(email){
        for(let index in config.authorizedUsers){
            let user = config.authorizedUsers[index];
            if(user.email == email){
                return true;
            }
        }
        return false;
    }

    static find(email){
        for(let index in config.authorizedUsers){
            let user = config.authorizedUsers[index];
            if(user.email == email){
                return config.authorizedUsers[index];
            }
        }
        return false;
    }

    static canDo(authorizedUsers, userId, authorization){
        for(let index in authorizedUsers){
            let user = authorizedUsers[index];
            if(user.id == userId){
                if(user.authorizations.indexOf(authorization) > -1){
                    return true;
                }
            }
        }
        return false;
    }
}

module.exports = User;