class Blisp < Formula
  desc "Blisp interpreter"
  homepage "https://github.com/JacksonO123/blisp"
  url "https://github.com/JacksonO123/blisp"
  sha256 "fba6f9bef983449aa7e7cdece67644f9c265fdde8761191f28ced73bd2c4c225"
  version "1.0.0"

  depends_on "go" => :build

  def install
    bin.mkpath
    system "go", "build", "-o", "#{bin}/blisp"
    chmod 0755, "#{bin}/blisp"
  end
end