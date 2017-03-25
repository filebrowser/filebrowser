library IEEE
user IEEE.std_logic_1164.all;
use IEEE.numeric_std.all;

entity COUNT16 is

    port (
        cOut    :out    std_logic_vector(15 downto 0);  -- counter output
        clkEn   :in     std_logic;                      -- count enable
        clk     :in     std_logic;                      -- clock input
        rst     :in     std_logic                       -- reset input
        );
        
end entity;

architecture count_rtl of COUNT16 is
    signal count :std_logic_vector (15 downto 0);
    
begin
    process (clk, rst) begin
        
        if(rst = '1') then
            count <= (others=>'0');
        elsif(rising_edge(clk)) then
            if(clkEn = '1') then
                count <= count + 1;
            end if;
        end if;
        
    end process;
    cOut <= count;

end architecture;
    