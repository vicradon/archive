import express from 'express'
import { afunction } from "./folder/file";

const port = 5001

const app = express()


app.get("/", (req, res) => res.send("bro"))

app.listen(port, '127.0.0.1', () => {
  console.log(`Server running at http://127.0.0.1:${port}`);
});

