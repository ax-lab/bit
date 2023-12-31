use super::*;

/// Spans reference a slice of text from a [`Source`].
#[derive(Default, Copy, Clone, Eq, PartialEq, Hash)]
pub struct Span<'a> {
	sta: usize,
	end: usize,
	ind: usize,
	src: Source<'a>,
}

pub trait HasSpan<'a> {
	fn span(&self) -> Span<'a>;
}

impl<'a> Span<'a> {
	pub(crate) fn new(src: Source<'a>, sta: usize, end: usize, ind: usize) -> Self {
		assert!(sta <= end && end <= src.len());
		Self { sta, end, ind, src }
	}

	pub fn empty() -> Self {
		Self::default()
	}

	pub fn range<T: IntoIterator<Item = U>, U: HasSpan<'a>>(elems: T) -> Span<'a> {
		let mut iter = elems.into_iter();
		if let Some(first) = iter.next() {
			let first = first.span();
			if let Some(last) = iter.last() {
				let last = last.span();
				Self::merge(first, last)
			} else {
				first
			}
		} else {
			Span::empty()
		}
	}

	pub fn display_text(&self) -> Option<Cow<'a, str>> {
		const MAX_LEN: usize = 30;

		let text = self.text();
		let text = if let Some(index) = text.find(|chr| chr == '\r' || chr == '\n') {
			&text[..index]
		} else {
			text
		};
		let text = text.trim_end();
		let suffix = if text.len() < self.len() { "…" } else { "" };
		let (prefix, text) = {
			let trimmed = text.trim_start();
			let prefix = if trimmed.len() < text.len() { "…" } else { "" };
			(prefix, trimmed)
		};

		let (text, suffix) = if let Some((n, _)) = text.char_indices().nth(MAX_LEN) {
			(&text[..n], "…")
		} else {
			(text, suffix)
		};

		if text.len() > 0 {
			let text = if suffix.len() + prefix.len() > 0 {
				let text = format!("{prefix}{text}{suffix}");
				Cow::Owned(text)
			} else {
				Cow::Borrowed(text)
			};
			Some(text)
		} else {
			None
		}
	}

	pub fn up_to(&self, other: Self) -> Self {
		assert_eq!(self.src, other.src);
		assert!(self.sta <= other.sta);
		Self {
			sta: self.sta,
			end: other.sta,
			ind: self.ind,
			src: self.src,
		}
	}

	pub fn to_end(&self) -> Self {
		Self {
			sta: self.end,
			end: self.end,
			ind: self.ind,
			src: self.src,
		}
	}

	pub fn truncated(&self, len: usize) -> Self {
		assert!(len <= self.len());
		Self {
			sta: self.sta,
			end: self.sta + len,
			ind: self.ind,
			src: self.src,
		}
	}

	pub fn merge(a: Self, b: Self) -> Self {
		if a.is_empty() {
			return b;
		}
		if b.is_empty() {
			return a;
		}

		assert_eq!(a.src, b.src);
		let (a, b) = if a.sta <= b.sta { (a, b) } else { (b, a) };
		Self {
			sta: a.sta,
			ind: a.ind,
			end: std::cmp::max(a.end, b.end),
			src: a.src,
		}
	}

	pub fn merged(self, other: Self) -> Self {
		Self::merge(self, other)
	}

	pub fn src(&self) -> Source<'a> {
		self.src
	}

	pub fn pos(&self) -> usize {
		self.sta
	}

	pub fn len(&self) -> usize {
		self.end - self.sta
	}

	pub fn end(&self) -> usize {
		self.end
	}

	pub fn indent(&self) -> usize {
		self.ind
	}

	pub fn text(&self) -> &'a str {
		unsafe { self.src.text().get_unchecked(self.pos()..self.end()) }
	}

	pub fn is_empty(&self) -> bool {
		self.sta == 0 && self.end == 0 && self.ind == 0 && self.src == Source::empty()
	}

	pub fn location(&self) -> Cursor<'a> {
		let mut cursor = Cursor::new(self.src());
		cursor.skip_len(self.sta);
		cursor
	}
}

impl<'a> Ord for Span<'a> {
	fn cmp(&self, other: &Self) -> Ordering {
		self.src
			.cmp(&other.src)
			.then_with(|| self.sta.cmp(&other.sta))
			.then_with(|| self.end.cmp(&other.end))
	}
}

impl<'a> PartialOrd for Span<'a> {
	fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
		Some(self.cmp(other))
	}
}

impl<'a> Display for Span<'a> {
	fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
		let loc = self.location();
		let len = self.len();
		write!(f, "{loc}")?;
		if len > 0 {
			write!(f, "+{len}")?;
		}
		Ok(())
	}
}

impl<'a> Debug for Span<'a> {
	fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
		let src = self.src();
		let pos = self.pos();
		let len = self.len();
		write!(f, "{src}:{pos}")?;
		if len > 0 {
			write!(f, "+{len}")?;
		}
		Ok(())
	}
}
