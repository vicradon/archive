// Download file
function downloadFile(filename) {
  window.location.href = `/download/${encodeURIComponent(filename)}`;
}

// Delete file with confirmation
async function deleteFile(filename) {
  if (!confirm(`Are you sure you want to delete "${filename}"?`)) {
    return;
  }

  try {
    const response = await fetch(`/delete/${encodeURIComponent(filename)}`, {
      method: 'DELETE'
    });

    if (response.ok) {
      // Remove the file item from the page
      const fileItems = document.querySelectorAll('.file-item');
      fileItems.forEach(item => {
        const nameElement = item.querySelector('.file-name');
        if (nameElement && nameElement.textContent === filename) {
          item.remove();
        }
      });

      // Check if there are no files left and show empty state
      const filesList = document.querySelector('.files-list');
      if (filesList && filesList.children.length === 0) {
        const container = document.querySelector('.container .card');
        container.innerHTML = `
          <div class="header-with-link">
            <h1>Uploaded Files</h1>
            <a href="/" class="view-uploads-link">Back to Upload</a>
          </div>
          <div class="empty-state">No uploaded files found</div>
        `;
      }
    } else {
      alert('Failed to delete file');
    }
  } catch (error) {
    console.error('Error deleting file:', error);
    alert('Error deleting file');
  }
}