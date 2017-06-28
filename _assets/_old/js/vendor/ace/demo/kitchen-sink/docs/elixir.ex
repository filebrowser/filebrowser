defmodule HelloModule do
  @moduledoc """
    This is supposed to be `markdown`.
    __Yes__ this is [mark](http://down.format)

    # Truly

    ## marked

    * with lists
    * more
    * and more

        Even.with(code)
        blocks |> with |> samples

    _Docs are first class citizens in Elixir_ (Jose Valim)
  """
  
  # A "Hello world" function
  def some_fun do
    IO.puts "Juhu Kinners!"
  end
  # A private function
  defp priv do
    is_regex ~r"""
       This is a regex
       spanning several
       lines.
    """
    x = elem({ :a, :b, :c }, 0)  #=> :a
  end
end

test_fun = fn(x) ->
  cond do
    x > 10 ->
      :greater_than_ten
    true ->
      :maybe_ten
  end
end