const { Client } = require('@elastic/elasticsearch')
const client = new Client({ node: 'http://localhost:9200' })

var esFunctions = {}

esFunctions.ping = function () {

client.cluster.health({}, function(error, resp) {
    if (error) {
        console.error('elasticsearch cluster is down!');
        console.log(error)
    } else {
        console.log(resp.body.status);
    }
});

}

esFunctions.checkExisting = function () {
    client.count({
        index: "alerts",
        q : ""
    })
}
module.exports = esFunctions