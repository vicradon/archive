const router = require("express").Router();
const messagingController = require("../controllers/messaging");
router.post("/:number", messagingController.send);

module.exports = router;
