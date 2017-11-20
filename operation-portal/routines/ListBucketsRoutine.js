const AWS = require('aws-sdk');
const Routine = require('../app/Routine');

class ListBucketsRoutine extends Routine {
    constructor(config){
        super(config);
    }

    process(successCallback, failCallback){
        let s3 = new AWS.S3();

        s3.listBuckets({}, function(err, data){
            if(err){
                failCallback(err);
            } else {
                successCallback(data.Buckets);
            }
        });
    }

    parseArgs(args) {
        return {};
    }

}

module.exports = ListBucketsRoutine;