from app.crawler import summarize, tags_from_text


def test_summarize_shortens_long_text():
    text = " ".join(["hello world."] * 80)
    assert len(summarize(text, 120)) <= 120


def test_tags_from_text_keeps_category():
    tags = tags_from_text("openai gpt", "OpenAI releases model", "ai")
    assert "openai" in tags
    assert "ai" in tags
