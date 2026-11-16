const mongo_client = require('mongodb').MongoClient;

const url = "mongodb://localhost:27017/mydb";

const mongo = mongo_client.connect(url, { useNewUrlParser: true, useUnifiedTopology: true }, (err, client) => {
  if (err) {
    console.error(err)
    return
  }
  const db = client.db('osinachi');

  const username = "Osinachi"

  console.log(`Database created by ${username}`);

})
