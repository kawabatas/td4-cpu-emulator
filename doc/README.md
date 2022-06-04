# レジスタ構成
4bitCPU
- 演算とデータ移動は4bit単位で行う

演算用のレジスタは A と B の2つ。ほぼ同等
- レジスタ A B はメモリ兼用、ワンチップマイコンなどの小規模なCPUに近い構成
- リセット直後はレジスタ A B ともに「0000」

プログラムカウンタも4bit
- よってプログラムは最大でも16ステップ
- リセットにより「0000」

フラグはC(キャリー)のみ
- 加算命令でキャリーが発生すると「1」になる
- 加算命令以外のすべての命令にも影響されるので、Cフラグを参照する命令は加算命令の直後に実行する

I/Oはともに4bit
- CPUにI/Oが内蔵されるところもワンチップマイコンに近い
- リセット直後の出力ポートは「0000」

TD4まとめ
- 汎用レジスタ 4bit*2
- アドレス空間 4bit(16バイト)
- プログラムカウンタ 4bit
- フラグレジスタ 1bit
- 算術演算 4bitの加算のみ

# 命令フォーマット
- すべての命令は8bitで構成される
- 上位4bitがオペレーションコード、下位4bitがイミディエイトデータ
- イミディエイトデータが不要な命令の場合には下位4bitを「0」で満たす必要がある

# 命令一覧
数値の転送命令
- MOV A,Im
  - イミディエイトデータをAレジスタへ転送
  - オペレーションコードは「0011」
- MOV B,Im
  - イミディエイトデータをBレジスタへ転送
  - オペレーションコードは「0111」

レジスタ間転送命令
- MOV B,A
  - BレジスタへAレジスタから転送（コピー）
  - オペレーションコードは「0100」、イミディエイトデータは不要
- MOV A,B
  - AレジスタへBレジスタから転送（コピー）
  - オペレーションコードは「0001」、イミディエイトデータは不要

加算命令
- ADD A,Im
  - Aレジスタにイミディエイトデータを加算
  - オペレーションコードは「0000」
- ADD B,Im
  - Bレジスタにイミディエイトデータを加算
  - オペレーションコードは「0101」

ジャンプ命令
- JMP Im
  - Im（ジャンプ先の番地）へジャンプ
  - オペレーションコードは「1111」

フラグと条件付きジャンプ（条件分岐）命令
- JNC Im
  - キャリーがなければジャンプ（Jump if Not Carry）
  - オペレーションコードは「1110」

入力命令(IN)
- IN A
  - 入力ポートをAレジスタに転送
  - オペレーションコードは「0010」、イミディエイトデータは不要
- IN B
  - 入力ポートをBレジスタに転送
  - オペレーションコードは「0110」、イミディエイトデータは不要

出力命令(OUT)
- OUT B
  - Bレジスタのデータを出力ポートへ転送
  - オペレーションコードは「1001」、イミディエイトデータは不要
- OUT Im
  - イミディエイトデータをそのまま出力ポートへ転送。手取り早くLEDをつけたり消したりに便利
  - オペレーションコードは「1011」

※NOP命令は特に用意していないが、`ADD A,0` が事実上のNOP