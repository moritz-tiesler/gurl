-- get_and_retrieve.lua

-- Global table to hold shortened URLs loaded from data generation
local shortened_urls = {}
local url_index = 0

-- Function to load shortened URLs from a file (if using Approach 2)
local function load_shortened_urls(filename)
   local file = io.open(filename, "r")
   if file then
      for line in file:lines() do
         table.insert(shortened_urls, line)
      end
      file:close()
      print("Loaded " .. #shortened_urls .. " shortened URLs.")
   else
      print("Error: Could not open " .. filename .. " for reading.")
   end
end

load_shortened_urls("bench/test_urls.txt")

if #shortened_urls == 0 then
   print("Error: No shortened URLs available for testing.")
   -- Returning false from init() might prevent the test from running
   return false
end

local num_urls = #shortened_urls

request = function()
   -- Cycle through the available shortened URLs
   url_index = url_index + 1
   if url_index > num_urls then
      url_index = 1 -- Loop back to the beginning
   end

   local shortened_code = shortened_urls[url_index]
   local get_url = "/url/" .. shortened_code -- Construct the GET URL

   wrk.method = "GET"
   -- wrk.headers["..."] -- Add any necessary headers for GET requests
   wrk.path = get_url
   return wrk.format(nil, get_url) -- Use wrk.request(url) to specify the URL
end
