const express = require("express");
const multer = require("multer");
const Docxtemplater = require("docxtemplater");
const PizZip = require("pizzip");
const path = require("path");
const fs = require("fs");

require("dotenv").config()

const app = express();
const upload = multer({ storage: multer.memoryStorage() });

app.use(express.static(path.join(__dirname, "public")));

/**
 * Wraps bare placeholder names in {curly braces} directly in the XML.
 * Word stores placeholders as plain text, so docxtemplater can't find them.
 */
function wrapPlaceholders(zip, placeholders) {
  const xmlParts = Object.keys(zip.files).filter(
    (name) => name.startsWith("word/") && name.endsWith(".xml")
  );
  for (const name of xmlParts) {
    let xml = zip.files[name].asText();
    let changed = false;
    for (const p of placeholders) {
      const escaped = p.replace(/[.*+?^${}()|[\]\\]/g, "\\$&").replace(/ +/g, "\\s+");
      const bare = new RegExp(`(?<!\\{)\\b${escaped}\\b(?!\\})`, "g");
      if (bare.test(xml)) {
        xml = xml.replace(bare, `{${p}}`);
        changed = true;
      }
    }
    if (changed) zip.file(name, xml);
  }
  return zip;
}

// ── Routes ────────────────────────────────────────────────────────────────────

/**
 * POST /preview
 * Body: multipart with field "docx"
 * Returns: { docxBase64: string } – the raw docx for client-side rendering
 */
app.post("/preview", upload.single("docx"), (req, res) => {
  try {
    if (!req.file) return res.status(400).json({ error: "No file uploaded." });
    res.json({ docxBase64: req.file.buffer.toString("base64") });
  } catch (err) {
    console.error("Preview error:", err);
    res.status(500).json({ error: err.message });
  }
});

/**
 * POST /replace
 * Body: multipart with "docx" file + "replacements" JSON string
 * Query:  ?preview=1  → returns JSON { filename, docxBase64 }
 * Default → streams the filled .docx as a file download
 */
app.post("/replace", upload.single("docx"), (req, res) => {
  try {
    if (!req.file) return res.status(400).json({ error: "No file uploaded." });

    let replacements = {};
    try {
      replacements = JSON.parse(req.body.replacements || "{}");
    } catch {
      return res.status(400).json({ error: "Invalid replacements JSON." });
    }

    const placeholders = Object.keys(replacements);
    if (placeholders.length === 0) {
      return res.status(400).json({ error: "No replacement fields provided." });
    }

    const content = req.file.buffer.toString("binary");
    let zip = new PizZip(content);
    zip = wrapPlaceholders(zip, placeholders);

    const doc = new Docxtemplater(zip, { paragraphLoop: true, linebreaks: true });
    doc.render(replacements);

    const outputBuffer = doc.toBuffer();
    const originalName = req.file.originalname.replace(/\.docx$/i, "");
    const filename = `${originalName}_filled.docx`;

    // Preview mode: return JSON with base64 docx for client-side rendering
    if (req.query.preview === "1") {
      return res.json({
        filename,
        docxBase64: outputBuffer.toString("base64"),
      });
    }

    // Default: stream docx download
    res.setHeader(
      "Content-Type",
      "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    );
    res.setHeader("Content-Disposition", `attachment; filename="${filename}"`);
    res.send(outputBuffer);
  } catch (err) {
    console.error("Replace error:", err);
    res.status(500).json({ error: err.message || "Failed to process document." });
  }
});

/**
 * GET /serve-file?path=/absolute/path/to/file.docx
 * Reads a .docx from the given filesystem path and streams it to the client.
 * Called by preview-file.html on every load so the page always reflects the
 * latest version of the file on disk.
 */
app.get("/serve-file", (req, res) => {
  const filePath = req.query.path;
  if (!filePath) return res.status(400).json({ error: "Missing ?path= query parameter." });
  if (!filePath.toLowerCase().endsWith(".docx")) return res.status(400).json({ error: "Only .docx files are supported." });
  if (!fs.existsSync(filePath)) return res.status(404).json({ error: "File not found: " + filePath });
  res.setHeader("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document");
  res.setHeader("Cache-Control", "no-store");
  fs.createReadStream(filePath).pipe(res);
});

const PORT = process.env.PORT || 4000;
app.listen(PORT, () => {
  console.log(`✅ Docx Replacer running → http://localhost:${PORT}`);
});