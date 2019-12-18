local function is_empty(s)
    return s == nil or s == ''
end

function do_test (ch, args)
	print(ch:name())
	print(ch:room())
	if not is_empty(args) then
	    print(args)
        ch:Send("Parsed " .. args .. ".\n")
	end
	m = TestList()
	print(m)
end
