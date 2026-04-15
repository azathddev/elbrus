const API_BASE_URL =
  window.location.protocol === "file:" ? "http://localhost:8081" : "";

const messageInput = document.getElementById("messageInput");
const audienceSelect = document.getElementById("audienceSelect");
const generateBtn = document.getElementById("generateBtn");
const gallery = document.getElementById("gallery");
const statusText = document.getElementById("statusText");

const generatedImages = [];

generateBtn.addEventListener("click", async () => {
  const message = messageInput.value.trim();
  const audience = audienceSelect.value.trim();

  if (!message) {
    showStatus("Введите текст рекламного сообщения", true);
    return;
  }

  setLoading(true);
  showStatus("Генерируем баннер через SDK...", false);

  try {
    const response = await fetch(`${API_BASE_URL}/api/generate`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ message, audience }),
    });

    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || "Ошибка генерации");
    }

    const result = data.result || {};
    let src = result.imageUrl || "";
    if (!src && result.imageBase64) {
      src = `data:image/png;base64,${result.imageBase64}`;
    }

    if (!src) {
      throw new Error("SDK не вернул изображение");
    }

    generatedImages.unshift({
      src,
      prompt: result.prompt || message,
      audience,
      createdAt: new Date().toLocaleString("ru-RU"),
    });

    renderGallery();
    showStatus("Изображение успешно сгенерировано", false);
  } catch (error) {
    showStatus(error.message, true);
  } finally {
    setLoading(false);
  }
});

function setLoading(isLoading) {
  generateBtn.disabled = isLoading;
  generateBtn.textContent = isLoading ? "Генерация..." : "Отправить";
}

function showStatus(text, isError) {
  statusText.textContent = text;
  statusText.classList.toggle("error", Boolean(isError));
}

function renderGallery() {
  if (!generatedImages.length) {
    gallery.innerHTML = "<p class='placeholder'>Пока нет изображений.</p>";
    return;
  }

  gallery.innerHTML = generatedImages
    .map(
      (item) => `
        <article class="card">
          <img src="${item.src}" alt="Сгенерированный баннер" />
          <div class="cardMeta">
            <p><strong>Аудитория:</strong> ${escapeHTML(item.audience)}</p>
            <p><strong>Время:</strong> ${escapeHTML(item.createdAt)}</p>
            <p><strong>Промпт:</strong> ${escapeHTML(item.prompt)}</p>
          </div>
        </article>
      `
    )
    .join("");
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

renderGallery();
