class Blisp < Formula
  desc "Blisp interpreter"
  homepage "https://github.com/JacksonO123/blisp"
  url "https://github.com/JacksonO123/blisp/releases/download/test/blisp-0-1-0.tar.gz"
  sha256 "e9af8adacc2112777e1289466250048337f012352e6519da8693231361304dcc"
  version "0.1.0"

  depends_on "go" => :build

  def install
    bin.mkpath
    system "go", "build", "-o", "#{bin}/blisp"
    chmod 0755, "#{bin}/blisp"
  end
end