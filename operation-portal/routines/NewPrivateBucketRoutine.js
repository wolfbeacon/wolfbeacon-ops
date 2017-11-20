const AWS = require('aws-sdk');
const Routine = require('../app/Routine');

class NewPrivateBucketRoutine extends Routine {
    constructor(config){
        super(config);
    }

    process(successCallback, failCallback){
        let params = {
            Bucket: this.config.bucketName
        };

        let s3 = new AWS.S3();

        s3.createBucket(params, function(err, data){
            if(err){
                fallCallback(err);
            } else {
                successCallback(data);
            }
        });
    }
}

module.exports = NewPrivateBucketRoutine;