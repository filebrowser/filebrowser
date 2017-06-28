always @(negedge reset or posedge clk) begin
  if (reset == 0) begin
    d_out <= 16'h0000;
    d_out_mem[resetcount] <= d_out;
    laststoredvalue <= d_out;
  end else begin
    d_out <= d_out + 1'b1; 
  end
end

always @(bufreadaddr)
  bufreadval = d_out_mem[bufreadaddr];