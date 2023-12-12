import express from "express"

const app = express()
const port = 4000

app.get('/', (req, res) => {
  res.send('Hello World!')
})

app.post("/login", (req, res)=> {
    console.log(req.body);
    res.status(200).end();
})

app.listen(port, () => {
  console.log(`Example app listening on port ${port}`)
})

