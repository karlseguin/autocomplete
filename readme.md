# AutoComplete
A trie written in Go optimized for autocomplete of english words/phrases.

The trie uses a coarse read-write mutex to ensure thread-safety.

Each node contains all possible matching ids, making retrieval fast (O(N) where N is the size
of the input) but not memory efficient. Nevertheless, 2800 video game titles
takes roughly 12MB. Normalization of input is optimized for english words. The sentence

    I Like <3Cats<3

will be indexed with the following three values:

    ilikecats
    likecats
    cats

This has proved a simple but effective approach.

Insert performance is nothing to write home about. Delete performance is poor (reducing the coarsness of the lock would help).

## Usage
You interact with the trie via the `Insert`, `Find` and `Remove` methods:

    titles := map[string]string {
      "16": "3d Body Adventure",
      "409": "Bit Bat",
      "796": "Double Dragon II The Revenge",
      "1455": "M1 Tank Platoon",
      "1776": "Panza Kick Boxing",
      "2021": "Romance Of The Three Kingdoms 3",
      "2455": "Terminal Velocity Cdrom",
      "2688": "Warriors Of Legend",
      "2802": "Yatzy",
    }
    
    //100 is the maximum title size we should worry about
    ac := autocomplete.New(100)
    for id, title := range titles {
      ac.Insert(id, title)
    }
    
    ids := ac.Find("ki")
    //ids == ["1776", "2021"] (order might be different)
    
    // remove sadly requires both the id and the title
    ac.Remove("2802")
