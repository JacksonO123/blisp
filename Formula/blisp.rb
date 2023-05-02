class Blisp < Formula
  desc "Blisp interpreter"
  homepage "https://github.com/JacksonO123/blisp"
  url "https://github.com/JacksonO123/blisp/archive/refs/tags/prod.tar.gz"
  sha256 "b4859e6a24847c4ae943331ec3e190ddc881ecd6d651a2eb25a3184f66e56f39"
  version "0.1.1"

  depends_on "go" => :build

  def install
    system "go", "build", "-o", "#{bin}/blisp"
    chmod 0755, "#{bin}/blisp"
  end
end