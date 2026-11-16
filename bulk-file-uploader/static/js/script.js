let selectedFiles = [];
let activeUploads = {};

// File type filters
const fileFilters = {
  all: "*/*",
  video: "video/*,.mov,.mkv",
  image: "image/*,.heic,.heif",
  document: ".pdf,.doc,.docx,.txt,.rtf",
  audio: "audio/*",
  archive: ".zip,.rar,.7z,.tar,.gz"
};

// Apply filter changes via radio buttons
document.addEventListener("DOMContentLoaded", () => {
  const fileInput = document.getElementById("file");
  const radios = document.querySelectorAll('input[name="fileFilter"]');

  radios.forEach((radio) => {
    radio.addEventListener("change", (e) => {
      const selected = e.target.value;
      fileInput.setAttribute("accept", fileFilters[selected]);
    });
  });

  // Default accept
  fileInput.setAttribute("accept", fileFilters.all);

  // File selection handler
  fileInput.addEventListener("change", handleFileSelect);
});

// Handle file selection
function handleFileSelect(e) {
  const newFiles = Array.from(e.target.files);
  selectedFiles = [...selectedFiles, ...newFiles];
  updateSelectedFilesList();
  e.target.value = "";
}

// Update the file list display
function updateSelectedFilesList() {
  const container = document.getElementById("selectedFiles");
  const uploadBtn = document.getElementById("uploadBtn");

  if (selectedFiles.length === 0) {
    container.classList.add("hidden");
    uploadBtn.disabled = true;
    return;
  }

  container.classList.remove("hidden");
  uploadBtn.disabled = false;

  container.innerHTML = selectedFiles
    .map(
      (file, index) => `
        <div class="file-item">
          <div class="file-info">
            <div class="file-name">${file.name}</div>
            <div class="file-size">${formatBytes(file.size)}</div>
          </div>
          <button class="remove-btn" onclick="removeFile(${index})">Remove</button>
        </div>
      `
    )
    .join("");
}

// Remove file from selection
function removeFile(index) {
  selectedFiles.splice(index, 1);
  updateSelectedFilesList();
}

// Human-readable size
function formatBytes(bytes) {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}

// Start upload
function startUploads() {
  if (selectedFiles.length === 0) return;

  const filesToUpload = [...selectedFiles];
  selectedFiles = [];
  updateSelectedFilesList();

  filesToUpload.forEach((file) => {
    const uploadId = Date.now() + "_" + Math.random();
    activeUploads[uploadId] = {
      file,
      paused: false,
      canceled: false,
      progress: 0,
    };

    addUploadToQueue(uploadId, file);
    uploadFileInChunks(uploadId, file);
  });
}

// Add upload to UI queue
function addUploadToQueue(uploadId, file) {
  const queueContainer = document.getElementById("uploadQueue");

  if (queueContainer.querySelector(".empty-state")) {
    queueContainer.innerHTML = "";
  }

  const uploadItem = document.createElement("div");
  uploadItem.className = "upload-item";
  uploadItem.id = "upload-" + uploadId;
  uploadItem.innerHTML = `
    <div class="upload-header">
      <div class="upload-name">${file.name}</div>
      <div class="upload-status" id="status-${uploadId}">0% • 0 MB/s</div>
    </div>
    <div class="progress-bar">
      <div class="progress-fill" id="progress-${uploadId}"></div>
    </div>
    <div class="upload-controls">
      <button class="btn-secondary" id="pause-${uploadId}" onclick="togglePause('${uploadId}')">Pause</button>
      <button class="btn-danger" onclick="cancelUpload('${uploadId}')">Cancel</button>
    </div>
  `;

  queueContainer.appendChild(uploadItem);
}

// Upload logic (chunked)
async function uploadFileInChunks(uploadId, file) {
  const chunkSize = 1024 * 1024; // 1MB
  const totalChunks = Math.ceil(file.size / chunkSize);
  let currentChunk = 0;
  let startTime = Date.now();
  let uploadedBytes = 0;

  while (currentChunk < totalChunks) {
    const upload = activeUploads[uploadId];
    if (!upload || upload.canceled) {
      await fetch("/cancel", { method: "POST" });
      removeUploadFromQueue(uploadId, "Canceled", true);
      return;
    }

    while (upload.paused) {
      await new Promise((resolve) => setTimeout(resolve, 100));
      if (!activeUploads[uploadId] || activeUploads[uploadId].canceled) {
        await fetch("/cancel", { method: "POST" });
        removeUploadFromQueue(uploadId, "Canceled", true);
        return;
      }
    }

    const offset = currentChunk * chunkSize;
    const start = offset;
    const end = Math.min(start + chunkSize, file.size);
    const chunk = file.slice(start, end);

    const formData = new FormData();
    formData.append("chunk", chunk);
    formData.append("filename", file.name);
    formData.append("chunkIndex", currentChunk);
    formData.append("totalChunks", totalChunks);
    formData.append("offset", offset.toString());
    formData.append("totalSize", file.size.toString());

    try {
      const response = await fetch("/upload-chunk", {
        method: "POST",
        body: formData,
      });

      if (!response.ok) throw new Error("Upload failed");
      const result = await response.json();

      if (result.status === "complete") {
        removeUploadFromQueue(uploadId, "Complete", false);
        return;
      }

      currentChunk++;
      uploadedBytes = Math.min(end, file.size);

      const elapsed = (Date.now() - startTime) / 1000;
      const speed = uploadedBytes / elapsed;
      const percent = Math.round((uploadedBytes / file.size) * 100);

      updateUploadProgress(uploadId, percent, speed);
    } catch (error) {
      console.error("Upload error:", error);
      removeUploadFromQueue(uploadId, "Failed: " + error.message, true);
      return;
    }
  }

  removeUploadFromQueue(uploadId, "Complete", false);
}

// Progress update
function updateUploadProgress(uploadId, percent, speed) {
  const progressBar = document.getElementById("progress-" + uploadId);
  const statusText = document.getElementById("status-" + uploadId);

  if (progressBar) progressBar.style.width = percent + "%";
  if (statusText) statusText.textContent = `${percent}% • ${formatBytes(speed)}/s`;

  if (activeUploads[uploadId]) {
    activeUploads[uploadId].progress = percent;
  }
}

// Pause/resume upload
function togglePause(uploadId) {
  const upload = activeUploads[uploadId];
  if (!upload) return;

  upload.paused = !upload.paused;
  const pauseBtn = document.getElementById("pause-" + uploadId);
  if (pauseBtn) pauseBtn.textContent = upload.paused ? "Resume" : "Pause";
}

// Cancel upload
function cancelUpload(uploadId) {
  if (activeUploads[uploadId]) {
    activeUploads[uploadId].canceled = true;
  }
}

// Remove upload from queue
function removeUploadFromQueue(uploadId, status, isError) {
  const uploadItem = document.getElementById("upload-" + uploadId);
  if (!uploadItem) return;

  const progressBar = uploadItem.querySelector(".progress-fill");
  const statusText = document.getElementById("status-" + uploadId);
  const controls = uploadItem.querySelector(".upload-controls");

  if (progressBar) {
    progressBar.classList.add(isError ? "error" : "complete");
    if (!isError) progressBar.style.width = "100%";
  }

  if (statusText) statusText.textContent = status;
  if (controls) controls.remove();

  delete activeUploads[uploadId];

  setTimeout(() => {
    uploadItem.remove();
    const queueContainer = document.getElementById("uploadQueue");
    if (queueContainer.children.length === 0) {
      queueContainer.innerHTML = '<div class="empty-state">No active uploads</div>';
    }
  }, 3000);
}
