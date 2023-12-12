import express from "express"

const app = express()
const port = 4000
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.get("/AuthPlayer", (req, res)=>{
    console.log(req.query);
    if (onlinePeople.has(req.query.token)) {
        console.log(onlinePeople.get(req.query.token))
        res.json(onlinePeople.get(req.query.token))
    }
    else {
        res.status(400).end();
    }
})
app.get('/', (req, res) => {
  res.send('Hello World!')
})

const onlinePeople = new Map();
app.post("/login", (req, res)=> {
    req.body.authToken = req.body.id;
    onlinePeople.set(req.body.authToken, req.body);
    res.json(req.body);
})


app.listen(port, () => {
  console.log(`Example app listening on port ${port}`)
})

