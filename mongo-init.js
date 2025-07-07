db = db.getSiblingDB("chatapp");

db.createUser({
  user: "chatapp",
  pwd: "chatapp123",
  roles: [{ role: "readWrite", db: "chatapp" }]
});
